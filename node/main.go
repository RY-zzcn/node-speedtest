package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"节点管理测速项目/node/api"
	"节点管理测速项目/node/config"
	"节点管理测速项目/node/speedtest"
)

var (
	configPath = flag.String("config", "./config.json", "配置文件路径")
	logPath    = flag.String("log", "./node.log", "日志文件路径")
	port       = flag.String("port", "3000", "节点监听端口")
	panelURL   = flag.String("panel", "", "面板URL")
	nodeID     = flag.String("id", "", "节点ID")
	nodeKey    = flag.String("key", "", "节点密钥")
)

// 节点状态
type NodeStatus struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Disk      float64   `json:"disk"`
	Uptime    int64     `json:"uptime"`
	Load      [3]float64 `json:"load"`
	NetworkRx int64     `json:"network_rx"`
	NetworkTx int64     `json:"network_tx"`
	Timestamp time.Time `json:"timestamp"`
}

// 全局变量
var (
	cfg              *config.Config
	speedTestManager *speedtest.SpeedTestManager
	lastNetStats     map[string]net.IOCountersStat
	lastNetStatsTime time.Time
)

func main() {
	flag.Parse()

	// 确保日志目录存在
	logDir := filepath.Dir(*logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("无法创建日志目录: %v", err)
	}

	// 设置日志输出
	logFile, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("无法打开日志文件: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// 加载配置
	config.SetConfigPath(*configPath)
	cfg = config.GetConfig()

	// 命令行参数覆盖配置文件
	if *port != "3000" {
		cfg.ListenPort = *port
	}
	if *panelURL != "" {
		cfg.PanelURL = *panelURL
	}
	if *nodeID != "" {
		cfg.NodeID = *nodeID
	}
	if *nodeKey != "" {
		cfg.NodeKey = *nodeKey
	}

	// 检查必要配置
	if cfg.PanelURL == "" {
		log.Fatal("未配置面板URL")
	}

	// 初始化网络统计
	initNetStats()

	// 如果未配置节点ID和密钥，尝试注册节点
	if cfg.NodeID == "" || cfg.NodeKey == "" {
		if err := registerNode(); err != nil {
			log.Fatalf("注册节点失败: %v", err)
		}
	}

	// 初始化测速管理器
	speedTestManager = speedtest.NewSpeedTestManager(cfg.PanelURL, cfg.NodeID, cfg.NodeKey)

	// 启动心跳协程
	go startHeartbeat()

	// 设置路由
	r := setupRouter()
	
	// 初始化API服务
	apiRouter := api.InitAPI()
	
	// 合并API路由到主路由
	r.Any("/api/*path", func(c *gin.Context) {
		apiRouter.HandleContext(c)
	})

	// 启动服务器
	log.Printf("节点服务启动在 http://localhost:%s", cfg.ListenPort)
	if err := r.Run(":" + cfg.ListenPort); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 设置路由
func setupRouter() *gin.Engine {
	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// 添加日志中间件
	r.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		log.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s",
			endTime.Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
		)
	})

	// 基本路由
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "节点服务运行中")
	})

	// 状态路由
	r.GET("/status", func(c *gin.Context) {
		status := getNodeStatus()
		c.JSON(http.StatusOK, status)
	})

	// 测速路由
	speedGroup := r.Group("/speedtest")
	{
		// 启动测速
		speedGroup.POST("/start", startSpeedTest)

		// 获取测速结果
		speedGroup.GET("/result/:id", getSpeedTestResult)

		// 获取所有测速结果
		speedGroup.GET("/results", getAllSpeedTestResults)

		// 测速文件下载
		speedGroup.GET("/download", func(c *gin.Context) {
			// 获取请求的文件大小
			sizeStr := c.DefaultQuery("size", "10")
			size := 10 // 默认10MB
			fmt.Sscanf(sizeStr, "%d", &size)
			if size <= 0 || size > 1000 {
				size = 10 // 限制最大1000MB
			}

			// 设置响应头
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=speedtest-%dmb.bin", size))
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

			// 生成并发送随机数据
			buf := make([]byte, 1024*1024) // 1MB缓冲区
			for i := 0; i < size; i++ {
				// 随机填充缓冲区
				for j := 0; j < len(buf); j += 64 {
					copy(buf[j:min(j+64, len(buf))], []byte(fmt.Sprintf("%064d", j)))
				}
				c.Writer.Write(buf)
				c.Writer.Flush()
			}
		})

		// 测速文件上传
		speedGroup.POST("/upload", func(c *gin.Context) {
			// 丢弃上传的数据
			_, err := io.Copy(ioutil.Discard, c.Request.Body)
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("上传失败: %v", err))
				return
			}
			c.String(http.StatusOK, "上传成功")
		})

		// Ping测试
		speedGroup.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
	}

	return r
}

// 获取节点状态
func getNodeStatus() NodeStatus {
	status := NodeStatus{
		ID:        cfg.NodeID,
		Timestamp: time.Now(),
	}

	// 获取主机名
	hostname, err := os.Hostname()
	if err == nil {
		status.Hostname = hostname
	}

	// 获取IP地址
	// 简化处理，实际应用中应该获取公网IP
	status.IP = "127.0.0.1"

	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		status.CPU = cpuPercent[0]
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		status.Memory = memInfo.UsedPercent
	}

	// 获取磁盘使用率
	diskInfo, err := disk.Usage("/")
	if err == nil {
		status.Disk = diskInfo.UsedPercent
	}

	// 获取系统负载
	if runtime.GOOS == "windows" {
		// Windows不支持获取系统负载，使用CPU使用率代替
		if len(cpuPercent) > 0 {
			status.Load = [3]float64{cpuPercent[0], cpuPercent[0], cpuPercent[0]}
		}
	} else {
		loadInfo, err := load.Avg()
		if err == nil {
			status.Load = [3]float64{loadInfo.Load1, loadInfo.Load5, loadInfo.Load15}
		}
	}

	// 获取系统运行时间
	hostInfo, err := host.Info()
	if err == nil {
		status.Uptime = int64(hostInfo.Uptime)
	}

	// 获取网络流量
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		// 计算网络流量速率
		if !lastNetStatsTime.IsZero() {
			duration := time.Since(lastNetStatsTime).Seconds()
			if duration > 0 {
				for _, stat := range netStats {
					if lastStat, ok := lastNetStats[stat.Name]; ok {
						status.NetworkRx += int64(float64(stat.BytesRecv-lastStat.BytesRecv) / duration)
						status.NetworkTx += int64(float64(stat.BytesSent-lastStat.BytesSent) / duration)
					}
				}
			}
		}

		// 更新上次统计数据
		updateNetStats(netStats)
	}

	return status
}

// 初始化网络统计
func initNetStats() {
	lastNetStats = make(map[string]net.IOCountersStat)
	netStats, err := net.IOCounters(false)
	if err == nil {
		updateNetStats(netStats)
	}
}

// 更新网络统计
func updateNetStats(netStats []net.IOCountersStat) {
	for _, stat := range netStats {
		lastNetStats[stat.Name] = stat
	}
	lastNetStatsTime = time.Now()
}

// 注册节点
func registerNode() error {
	log.Println("开始注册节点...")

	// 获取主机信息
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}

	// 准备注册请求
	reqBody := map[string]interface{}{
		"name":     hostname,
		"ip":       "127.0.0.1", // 简化处理，实际应该获取公网IP
		"location": "未知位置",
		"tags":     []string{"自动注册"},
		"version":  "1.0.0",
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/api/node/register", cfg.PanelURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("执行HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("注册失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID        string `json:"id"`
			SecretKey string `json:"secretKey"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应码
	if response.Code != 0 {
		return fmt.Errorf("注册失败: %s", response.Message)
	}

	// 保存节点ID和密钥
	cfg.NodeID = response.Data.ID
	cfg.NodeKey = response.Data.SecretKey

	// 保存配置
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	log.Printf("节点注册成功，ID: %s", cfg.NodeID)
	return nil
}

// 启动心跳
func startHeartbeat() {
	for {
		// 获取节点状态
		status := getNodeStatus()

		// 准备心跳请求
		reqBody := map[string]interface{}{
			"id":         status.ID,
			"cpu":        status.CPU,
			"memory":     status.Memory,
			"disk":       status.Disk,
			"uptime":     status.Uptime,
			"load":       status.Load,
			"network_rx": status.NetworkRx,
			"network_tx": status.NetworkTx,
			"timestamp":  status.Timestamp,
		}

		// 序列化请求体
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Printf("序列化心跳请求失败: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		// 创建HTTP请求
		url := fmt.Sprintf("%s/api/node/heartbeat", cfg.PanelURL)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("创建心跳请求失败: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Node-Key", cfg.NodeKey)

		// 执行请求
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("发送心跳失败: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		// 读取响应内容
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("读取心跳响应失败: %v", err)
		} else if resp.StatusCode != http.StatusOK {
			log.Printf("心跳响应异常，状态码: %d, 响应: %s", resp.StatusCode, string(body))
		} else {
			log.Println("心跳发送成功")
		}

		// 等待下一次心跳
		time.Sleep(30 * time.Second)
	}
}

// 启动测速
func startSpeedTest(c *gin.Context) {
	var req speedtest.SpeedTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("无效的请求数据: %v", err),
		})
		return
	}

	// 检查必要字段
	if req.ID == "" || req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要字段",
		})
		return
	}

	// 如果未指定源节点ID，使用本节点ID
	if req.SourceNodeID == "" {
		req.SourceNodeID = cfg.NodeID
	}

	// 启动测速
	result, err := speedTestManager.StartTest(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("启动测速失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "测速已启动",
		"id":      result.ID,
	})
}

// 获取测速结果
func getSpeedTestResult(c *gin.Context) {
	testID := c.Param("id")
	result, exists := speedTestManager.GetTestResult(testID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "测速结果不存在",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// 获取所有测速结果
func getAllSpeedTestResults(c *gin.Context) {
	results := speedTestManager.GetAllTests()
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(results),
	})
}

// 辅助函数：取最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 