package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/static"

	"./api"
	"./models"
	"./config"
)

var (
	port     = flag.String("port", "8080", "面板监听端口")
	dbPath   = flag.String("db", "./data.db", "数据库路径")
	logPath  = flag.String("log", "./panel.log", "日志文件路径")
	confPath = flag.String("conf", "./config.json", "配置文件路径")
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

	// 初始化配置
	config.SetConfigPath(*confPath)
	cfg := config.GetConfig()
	
	// 如果命令行指定了端口，覆盖配置文件中的端口
	if *port != "8080" {
		cfg.ListenPort = *port
	}
	
	// 如果命令行指定了数据库路径，覆盖配置文件中的路径
	if *dbPath != "./data.db" {
		cfg.DatabasePath = *dbPath
	}

	// 初始化数据库
	if err := models.InitDB(cfg.DatabasePath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 启动节点状态检查协程
	go checkNodesStatus()

	// 初始化路由
	r := setupRouter()

	// 启动服务器
	log.Printf("面板服务启动在 http://localhost:%s", cfg.ListenPort)
	if err := r.Run(":" + cfg.ListenPort); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func setupRouter() *gin.Engine {
	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 使用中间件
	r.Use(api.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(api.CORSMiddleware())

	// 静态文件服务
	r.Use(static.Serve("/", static.LocalFile("./web", false)))

	// API路由组
	apiGroup := r.Group("/api")
	{
		// 公开API
		apiGroup.POST("/login", api.LoginHandler)
		apiGroup.POST("/logout", api.LogoutHandler)
		
		// 节点安装脚本和下载API（公开）
		apiGroup.GET("/install.sh", api.GetInstallScriptHandler)
		apiGroup.GET("/download/:arch", api.DownloadNodeHandler)
		
		// 节点注册API（公开）
		apiGroup.POST("/node/register", api.RegisterNodeHandler)
		
		// 需要节点认证的API
		nodeGroup := apiGroup.Group("/node")
		nodeGroup.Use(api.NodeAuthMiddleware())
		{
			nodeGroup.POST("/heartbeat", api.NodeHeartbeatHandler)
			nodeGroup.POST("/speedtest/result", api.UpdateSpeedTestResultHandler)
		}
		
		// 需要用户认证的API
		authGroup := apiGroup.Group("")
		authGroup.Use(api.AuthMiddleware())
		{
			// 用户API
			authGroup.GET("/user", api.GetCurrentUserHandler)
			
			// 节点管理API
			authGroup.GET("/nodes", api.GetNodesHandler)
			authGroup.GET("/nodes/:id", api.GetNodeHandler)
			authGroup.PUT("/nodes/:id", api.UpdateNodeHandler)
			authGroup.DELETE("/nodes/:id", api.DeleteNodeHandler)
			authGroup.GET("/nodes/:id/install-command", api.GenerateInstallCommandHandler)
			
			// 测速API
			authGroup.POST("/speedtest", api.StartSpeedTestHandler)
			authGroup.GET("/speedtest/results", api.GetSpeedTestResultsHandler)
			authGroup.GET("/speedtest/results/:id", api.GetSpeedTestResultHandler)
			
			// 系统设置API
			authGroup.GET("/settings", api.GetSettingsHandler)
			authGroup.PUT("/settings", api.UpdateSettingsHandler)
			
			// 系统统计API
			authGroup.GET("/stats", api.GetStatsHandler)
		}
	}

	return r
}

// 检查节点状态
func checkNodesStatus() {
	for {
		// 获取节点超时设置
		nodeTimeout := 60 // 默认60秒
		nodeTimeoutStr, err := models.GetSetting("node_timeout")
		if err == nil && nodeTimeoutStr != "" {
			if nt, err := strconv.Atoi(nodeTimeoutStr); err == nil {
				nodeTimeout = nt
			}
		}

		// 获取所有节点
		nodes, err := models.GetAllNodes()
		if err != nil {
			log.Printf("获取节点列表失败: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		// 检查每个节点的状态
		for _, node := range nodes {
			// 如果节点最后心跳时间超过超时时间，标记为离线
			if node.Status == models.NodeStatusOnline && time.Since(node.LastSeen) > time.Duration(nodeTimeout)*time.Second {
				log.Printf("节点 %s (%s) 超时，标记为离线", node.Name, node.ID)
				if err := models.UpdateNodeStatus(node.ID, models.NodeStatusOffline); err != nil {
					log.Printf("更新节点状态失败: %v", err)
				}
			}
		}

		// 每30秒检查一次
		time.Sleep(30 * time.Second)
	}
} 