package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// 配置结构体
type Config struct {
	ListenPort        string `json:"listen_port"`
	DatabasePath      string `json:"database_path"`
	LogPath           string `json:"log_path"`
	SecretKey         string `json:"secret_key"`
	AdminUsername     string `json:"admin_username"`
	AdminPassword     string `json:"admin_password"`
	PanelURL          string `json:"panel_url"`
	NodeTimeout       int    `json:"node_timeout"`
	NodeCheckInterval int    `json:"node_check_interval"`
	SpeedtestTimeout  int    `json:"speedtest_timeout"`
	MaxConcurrentTests int   `json:"max_concurrent_tests"`
	GithubRepo        string `json:"github_repo"`
	GithubVersion     string `json:"github_version"`
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

// 主函数
func main() {
	fmt.Println("节点管理测速系统 - 面板服务")
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

	log.Println("面板服务启动")
	log.Printf("配置加载成功，监听端口: %s", config.ListenPort)

	// 设置HTTP路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "节点管理测速系统 - 面板服务正在运行")
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
	fmt.Printf("面板服务启动，监听地址: http://localhost%s\n", serverAddr)
	log.Printf("面板服务启动，监听地址: http://localhost%s", serverAddr)
	
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
} 