package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// 配置结构体
type Config struct {
	ListenPort        string `json:"listen_port"`
	LogPath           string `json:"log_path"`
	PanelURL          string `json:"panel_url"`
	NodeID            string `json:"node_id"`
	NodeKey           string `json:"node_key"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
	DownloadThreads   int    `json:"download_threads"`
	UploadThreads     int    `json:"upload_threads"`
	PingCount         int    `json:"ping_count"`
}

// 全局配置变量
var config Config

// 加载配置文件
func loadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&config)
}

// 初始化日志
func initLogger(logPath string) (*os.File, error) {
	// 确保日志目录存在
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// 打开日志文件
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// 设置日志输出
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logFile, nil
}

// 发送心跳包到面板
func sendHeartbeat() {
	if config.PanelURL == "" || config.NodeID == "" || config.NodeKey == "" {
		log.Println("面板URL、节点ID或节点密钥未设置，无法发送心跳")
		return
	}

	heartbeatURL := fmt.Sprintf("%s/api/node/heartbeat", config.PanelURL)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 构建请求
	req, err := http.NewRequest("POST", heartbeatURL, nil)
	if err != nil {
		log.Printf("创建心跳请求失败: %v", err)
		return
	}

	// 添加认证头
	req.Header.Set("Node-ID", config.NodeID)
	req.Header.Set("Node-Key", config.NodeKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送心跳失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("心跳响应状态码异常: %d", resp.StatusCode)
		return
	}

	log.Println("心跳发送成功")
}

// 启动心跳定时任务
func startHeartbeatTask() {
	interval := time.Duration(config.HeartbeatInterval) * time.Second
	if interval < 10*time.Second {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				sendHeartbeat()
			}
		}
	}()

	log.Printf("心跳任务启动，间隔: %v", interval)
}

// 主函数
func main() {
	fmt.Println("节点管理测速系统 - 节点服务")
	fmt.Println("版本: v1.0.0")

	// 解析命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	// 加载配置
	if err := loadConfig(*configPath); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logFile, err := initLogger(config.LogPath)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	log.Println("节点服务启动")
	log.Printf("配置加载成功，监听端口: %s", config.ListenPort)

	// 启动心跳任务
	startHeartbeatTask()

	// 设置HTTP路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "节点管理测速系统 - 节点服务正在运行")
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "running",
			"version": "v1.0.0",
		})
	})

	// 启动HTTP服务器
	serverAddr := ":" + config.ListenPort
	fmt.Printf("节点服务启动，监听地址: http://localhost%s\n", serverAddr)
	log.Printf("节点服务启动，监听地址: http://localhost%s", serverAddr)
	
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
} 