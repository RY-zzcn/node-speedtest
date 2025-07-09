package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"节点管理测速项目/node/config"
	"节点管理测速项目/node/speedtest"
)

// 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 测速任务请求
type SpeedtestRequest struct {
	TargetURL    string `json:"target_url"`
	Type         string `json:"type"`         // "download", "upload", "ping", "full"
	Size         int    `json:"size"`         // 测试文件大小(MB)
	Timeout      int    `json:"timeout"`      // 超时时间(秒)
	TaskID       string `json:"task_id"`      // 任务ID
	CallbackURL  string `json:"callback_url"` // 回调URL
}

// 节点状态
type NodeStatus struct {
	NodeName    string    `json:"node_name"`
	Version     string    `json:"version"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	Uptime      int64     `json:"uptime"`
	StartTime   time.Time `json:"start_time"`
}

var (
	router      *gin.Engine
	startTime   time.Time
	apiVersion  = "1.0.0"
	nodeStatus  NodeStatus
)

// 初始化API服务
func InitAPI() *gin.Engine {
	// 记录启动时间
	startTime = time.Now()

	// 创建Gin路由
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware())
	router.Use(AuthMiddleware())

	// 注册路由
	router.GET("/api/status", handleStatus)
	router.POST("/api/speedtest", handleSpeedtest)
	router.GET("/api/speedtest/:task_id", handleGetSpeedtestResult)
	router.POST("/api/config", handleUpdateConfig)
	router.GET("/api/config", handleGetConfig)

	// 更新节点状态
	updateNodeStatus()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			updateNodeStatus()
		}
	}()

	return router
}

// 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		log.Printf("[API] %s %s %d %s", method, path, statusCode, latency)
	}
}

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取节点密钥
		nodeKey := config.GetConfig().NodeKey

		// 检查请求头中的密钥
		authKey := c.GetHeader("X-Node-Key")
		if authKey == "" {
			// 也接受查询参数中的密钥
			authKey = c.Query("node_key")
		}

		// 验证密钥
		if authKey != nodeKey {
			c.JSON(http.StatusUnauthorized, Response{
				Code:    401,
				Message: "未授权访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 处理获取节点状态请求
func handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "成功",
		Data:    nodeStatus,
	})
}

// 处理测速请求
func handleSpeedtest(c *gin.Context) {
	var req SpeedtestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的请求参数",
		})
		return
	}

	// 设置默认值
	if req.Timeout <= 0 {
		req.Timeout = config.GetConfig().SpeedtestTimeout
	}
	if req.Size <= 0 {
		req.Size = 100 // 默认100MB
	}

	// 创建测速任务
	go func() {
		var result *speedtest.Result
		var err error

		switch req.Type {
		case "download":
			result, err = speedtest.TestDownload(req.TargetURL, req.Size, req.Timeout)
		case "upload":
			result, err = speedtest.TestUpload(req.TargetURL, req.Size, req.Timeout)
		case "ping":
			result, err = speedtest.TestPing(req.TargetURL, config.GetConfig().PingCount)
		case "full":
			result, err = speedtest.TestFull(req.TargetURL, req.Size, req.Timeout)
		default:
			log.Printf("未知的测速类型: %s", req.Type)
			return
		}

		// 测速完成后回调
		if req.CallbackURL != "" {
			sendCallback(req.CallbackURL, req.TaskID, result, err)
		}
	}()

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "测速任务已创建",
		Data: map[string]string{
			"task_id": req.TaskID,
		},
	})
}

// 处理获取测速结果请求
func handleGetSpeedtestResult(c *gin.Context) {
	taskID := c.Param("task_id")
	
	// 这里应该从存储中获取测速结果
	// 由于我们是异步执行测速任务，需要一个存储机制来保存结果
	// 这里简化处理，返回任务不存在
	c.JSON(http.StatusNotFound, Response{
		Code:    404,
		Message: "测速任务不存在或尚未完成",
	})
}

// 处理更新配置请求
func handleUpdateConfig(c *gin.Context) {
	var cfg config.Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的配置参数",
		})
		return
	}

	// 更新配置
	config.UpdateConfig(cfg)

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "配置已更新",
	})
}

// 处理获取配置请求
func handleGetConfig(c *gin.Context) {
	cfg := config.GetConfig()

	// 敏感信息处理
	cfgCopy := *cfg
	cfgCopy.NodeKey = "******" // 隐藏节点密钥

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "成功",
		Data:    cfgCopy,
	})
}

// 发送测速结果回调
func sendCallback(callbackURL string, taskID string, result *speedtest.Result, err error) {
	// 准备回调数据
	callbackData := map[string]interface{}{
		"task_id": taskID,
		"success": err == nil,
	}

	if err != nil {
		callbackData["error"] = err.Error()
	} else if result != nil {
		callbackData["result"] = result
	}

	// 序列化数据
	jsonData, err := json.Marshal(callbackData)
	if err != nil {
		log.Printf("序列化回调数据失败: %v", err)
		return
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", callbackURL, nil)
	if err != nil {
		log.Printf("创建回调请求失败: %v", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Key", config.GetConfig().NodeKey)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送回调请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("测速结果回调发送成功: %s, 状态码: %d", callbackURL, resp.StatusCode)
}

// 更新节点状态
func updateNodeStatus() {
	// 这里应该实现获取系统资源使用情况的逻辑
	// 简化处理，使用模拟数据
	nodeStatus = NodeStatus{
		NodeName:    config.GetConfig().NodeName,
		Version:     apiVersion,
		CPUUsage:    30.5,
		MemoryUsage: 45.2,
		DiskUsage:   60.8,
		Uptime:      int64(time.Since(startTime).Seconds()),
		StartTime:   startTime,
	}
} 