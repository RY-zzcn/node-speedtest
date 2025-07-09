package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"../auth"
	"../config"
	"../models"
)

// 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// API错误
func APIError(c *gin.Context, err error) {
	log.Printf("API错误: %v", err)
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: fmt.Sprintf("服务器错误: %v", err),
	})
}

// 需要登录
func RequireLogin(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: "需要登录",
	})
	c.Abort()
}

// 节点API处理函数

// 获取所有节点
func GetNodesHandler(c *gin.Context) {
	nodes, err := models.GetAllNodes()
	if err != nil {
		APIError(c, err)
		return
	}

	SuccessResponse(c, gin.H{
		"nodes": nodes,
		"total": len(nodes),
	})
}

// 获取单个节点
func GetNodeHandler(c *gin.Context) {
	nodeID := c.Param("id")
	node, err := models.GetNode(nodeID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("节点不存在: %s", nodeID))
		return
	}

	SuccessResponse(c, node)
}

// 创建/注册节点
func RegisterNodeHandler(c *gin.Context) {
	var req models.NodeRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}

	// 创建新节点
	node := &models.Node{
		ID:          uuid.New().String(),
		Name:        req.Name,
		IP:          req.IP,
		Location:    req.Location,
		Status:      models.NodeStatusOffline,
		LastSeen:    time.Now(),
		CreatedAt:   time.Now(),
		Description: req.Description,
		Tags:        req.Tags,
		Version:     req.Version,
	}

	// 保存节点
	if err := models.SaveNode(node); err != nil {
		APIError(c, err)
		return
	}

	// 生成节点密钥
	secretKey := fmt.Sprintf("sk_%s_%d", node.ID, time.Now().Unix())

	// 返回节点ID和密钥
	SuccessResponse(c, models.NodeRegisterResponse{
		ID:        node.ID,
		SecretKey: secretKey,
	})
}

// 更新节点
func UpdateNodeHandler(c *gin.Context) {
	nodeID := c.Param("id")
	
	// 检查节点是否存在
	existingNode, err := models.GetNode(nodeID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("节点不存在: %s", nodeID))
		return
	}

	var req models.Node
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}

	// 更新节点信息
	existingNode.Name = req.Name
	existingNode.IP = req.IP
	existingNode.Location = req.Location
	existingNode.Description = req.Description
	existingNode.Tags = req.Tags

	// 保存更新后的节点
	if err := models.SaveNode(existingNode); err != nil {
		APIError(c, err)
		return
	}

	SuccessResponse(c, existingNode)
}

// 删除节点
func DeleteNodeHandler(c *gin.Context) {
	nodeID := c.Param("id")
	
	// 检查节点是否存在
	_, err := models.GetNode(nodeID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("节点不存在: %s", nodeID))
		return
	}

	// 删除节点
	if err := models.DeleteNode(nodeID); err != nil {
		APIError(c, err)
		return
	}

	SuccessResponse(c, gin.H{"message": "节点已删除"})
}

// 节点心跳
func NodeHeartbeatHandler(c *gin.Context) {
	var heartbeat models.NodeHeartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的心跳数据: %v", err))
		return
	}

	// 设置心跳时间
	heartbeat.Timestamp = time.Now()

	// 更新节点心跳
	if err := models.UpdateNodeHeartbeat(&heartbeat); err != nil {
		APIError(c, err)
		return
	}

	SuccessResponse(c, gin.H{"message": "心跳更新成功"})
}

// 测速API处理函数

// 开始测速
func StartSpeedTestHandler(c *gin.Context) {
	var req models.SpeedTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}

	// 检查源节点和目标节点是否存在
	_, err := models.GetNode(req.SourceNodeID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("源节点不存在: %s", req.SourceNodeID))
		return
	}

	_, err = models.GetNode(req.TargetNodeID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("目标节点不存在: %s", req.TargetNodeID))
		return
	}

	// 创建测速结果
	result := &models.SpeedTestResult{
		ID:           uuid.New().String(),
		SourceNodeID: req.SourceNodeID,
		TargetNodeID: req.TargetNodeID,
		Type:         req.Type,
		Status:       models.SpeedTestStatusPending,
		StartTime:    time.Now(),
	}

	// 保存测速结果
	if err := models.SaveSpeedTestResult(result); err != nil {
		APIError(c, err)
		return
	}

	// TODO: 向源节点发送测速请求

	SuccessResponse(c, gin.H{
		"id":      result.ID,
		"message": "测速任务已创建",
	})
}

// 获取测速结果
func GetSpeedTestResultHandler(c *gin.Context) {
	resultID := c.Param("id")
	result, err := models.GetSpeedTestResult(resultID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("测速结果不存在: %s", resultID))
		return
	}

	SuccessResponse(c, result)
}

// 获取所有测速结果
func GetSpeedTestResultsHandler(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	
	// 获取过滤参数
	nodeID := c.Query("nodeId")
	
	var results []models.SpeedTestResult
	var err error
	
	if nodeID != "" {
		// 获取指定节点的测速结果
		results, err = models.GetNodeSpeedTestResults(nodeID)
	} else {
		// 获取所有测速结果
		results, err = models.GetAllSpeedTestResults()
	}
	
	if err != nil {
		APIError(c, err)
		return
	}
	
	// 简单的分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(results) {
		start = 0
		end = 0
	}
	if end > len(results) {
		end = len(results)
	}
	
	pagedResults := results
	if start < len(results) {
		pagedResults = results[start:end]
	} else {
		pagedResults = []models.SpeedTestResult{}
	}
	
	SuccessResponse(c, gin.H{
		"results": pagedResults,
		"total":   len(results),
		"page":    page,
		"pageSize": pageSize,
	})
}

// 更新测速结果
func UpdateSpeedTestResultHandler(c *gin.Context) {
	resultID := c.Param("id")
	
	// 检查测速结果是否存在
	existingResult, err := models.GetSpeedTestResult(resultID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("测速结果不存在: %s", resultID))
		return
	}
	
	var req models.SpeedTestResult
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}
	
	// 更新测速结果
	existingResult.Status = req.Status
	existingResult.EndTime = req.EndTime
	existingResult.Duration = req.Duration
	existingResult.DownloadSpeed = req.DownloadSpeed
	existingResult.UploadSpeed = req.UploadSpeed
	existingResult.Ping = req.Ping
	existingResult.Jitter = req.Jitter
	existingResult.PacketLoss = req.PacketLoss
	existingResult.ErrorMessage = req.ErrorMessage
	
	// 保存更新后的测速结果
	if err := models.SaveSpeedTestResult(existingResult); err != nil {
		APIError(c, err)
		return
	}
	
	SuccessResponse(c, existingResult)
}

// 用户API处理函数

// 用户登录
func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}
	
	// 验证用户登录
	valid, userID, err := models.ValidateUser(req.Username, req.Password)
	if err != nil {
		APIError(c, err)
		return
	}
	
	if !valid {
		ErrorResponse(c, 401, "用户名或密码错误")
		return
	}
	
	// 获取用户信息
	user, err := models.GetUser(userID)
	if err != nil {
		APIError(c, err)
		return
	}
	
	// 生成JWT令牌
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		APIError(c, err)
		return
	}
	
	// 设置会话Cookie
	c.SetCookie("session_token", token, 86400, "/", "", false, true)
	
	SuccessResponse(c, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// 用户注册
func RegisterHandler(c *gin.Context) {
	// 只允许管理员创建新用户
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		ErrorResponse(c, 403, "只有管理员可以创建新用户")
		return
	}
	
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Role     string `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}
	
	// 检查用户名是否已存在
	exists, err := models.UserExists(req.Username)
	if err != nil {
		APIError(c, err)
		return
	}
	
	if exists {
		ErrorResponse(c, 400, "用户名已存在")
		return
	}
	
	// 哈希密码
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		APIError(c, err)
		return
	}
	
	// 创建用户
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		Role:         req.Role,
		CreatedAt:    time.Now(),
	}
	
	// 保存用户
	if err := models.SaveUser(user); err != nil {
		APIError(c, err)
		return
	}
	
	SuccessResponse(c, gin.H{
		"message": "用户创建成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// 用户登出
func LogoutHandler(c *gin.Context) {
	// 清除会话Cookie
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	
	SuccessResponse(c, gin.H{
		"message": "登出成功",
	})
}

// 获取当前用户信息
func GetCurrentUserHandler(c *gin.Context) {
	// 从上下文中获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		RequireLogin(c)
		return
	}
	
	username, _ := c.Get("username")
	userRole, _ := c.Get("userRole")
	
	SuccessResponse(c, gin.H{
		"id":       userID,
		"username": username,
		"role":     userRole,
	})
}

// 系统设置API处理函数

// 获取系统设置
func GetSettingsHandler(c *gin.Context) {
	// 获取所有系统设置
	settings := map[string]string{}
	
	// 获取监听端口设置
	listenPort, err := models.GetSetting("listen_port")
	if err != nil {
		APIError(c, err)
		return
	}
	if listenPort != "" {
		settings["listen_port"] = listenPort
	} else {
		settings["listen_port"] = "8080" // 默认值
	}
	
	// 获取节点超时设置
	nodeTimeout, err := models.GetSetting("node_timeout")
	if err != nil {
		APIError(c, err)
		return
	}
	if nodeTimeout != "" {
		settings["node_timeout"] = nodeTimeout
	} else {
		settings["node_timeout"] = "60" // 默认值
	}
	
	// 获取节点检查间隔设置
	nodeCheckInterval, err := models.GetSetting("node_check_interval")
	if err != nil {
		APIError(c, err)
		return
	}
	if nodeCheckInterval != "" {
		settings["node_check_interval"] = nodeCheckInterval
	} else {
		settings["node_check_interval"] = "30" // 默认值
	}
	
	// 获取测速超时设置
	speedtestTimeout, err := models.GetSetting("speedtest_timeout")
	if err != nil {
		APIError(c, err)
		return
	}
	if speedtestTimeout != "" {
		settings["speedtest_timeout"] = speedtestTimeout
	} else {
		settings["speedtest_timeout"] = "120" // 默认值
	}
	
	// 获取最大并发测试数设置
	maxConcurrentTests, err := models.GetSetting("max_concurrent_tests")
	if err != nil {
		APIError(c, err)
		return
	}
	if maxConcurrentTests != "" {
		settings["max_concurrent_tests"] = maxConcurrentTests
	} else {
		settings["max_concurrent_tests"] = "3" // 默认值
	}
	
	SuccessResponse(c, settings)
}

// 更新系统设置
func UpdateSettingsHandler(c *gin.Context) {
	var settings map[string]string
	if err := c.ShouldBindJSON(&settings); err != nil {
		ErrorResponse(c, 400, fmt.Sprintf("无效的请求数据: %v", err))
		return
	}
	
	// 更新设置
	for key, value := range settings {
		if err := models.SaveSetting(key, value); err != nil {
			APIError(c, err)
			return
		}
	}
	
	SuccessResponse(c, gin.H{
		"message": "设置已更新",
	})
}

// 系统统计API处理函数

// 获取系统统计信息
func GetStatsHandler(c *gin.Context) {
	// 获取所有节点
	nodes, err := models.GetAllNodes()
	if err != nil {
		APIError(c, err)
		return
	}
	
	// 统计在线和离线节点数量
	onlineNodes := 0
	offlineNodes := 0
	for _, node := range nodes {
		if node.Status == models.NodeStatusOnline {
			onlineNodes++
		} else {
			offlineNodes++
		}
	}
	
	// 获取所有测速结果
	results, err := models.GetAllSpeedTestResults()
	if err != nil {
		APIError(c, err)
		return
	}
	
	// 统计今日测速次数
	todayTests := 0
	today := time.Now().Truncate(24 * time.Hour)
	for _, result := range results {
		if result.StartTime.After(today) {
			todayTests++
		}
	}
	
	// TODO: 获取面板服务器的系统信息
	
	SuccessResponse(c, gin.H{
		"onlineNodes":  onlineNodes,
		"offlineNodes": offlineNodes,
		"totalNodes":   len(nodes),
		"todayTests":   todayTests,
		"totalTests":   len(results),
		"cpuUsage":     30, // 示例值，应该从实际系统获取
		"memoryUsage":  40, // 示例值，应该从实际系统获取
		"diskUsage":    50, // 示例值，应该从实际系统获取
	})
} 

// 安装脚本和节点下载API处理函数

// 获取节点安装脚本
func GetInstallScriptHandler(c *gin.Context) {
	// 设置响应头
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=install.sh")

	// 获取面板URL
	panelURL := config.GetConfig().PanelURL
	if panelURL == "" {
		// 如果配置中没有设置面板URL，则使用请求中的Host
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		panelURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 读取安装脚本模板
	templatePath := "./web/install_template.sh"
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Printf("读取安装脚本模板失败: %v", err)
		c.String(http.StatusInternalServerError, "读取安装脚本模板失败")
		return
	}

	// 解析模板
	tmpl, err := template.New("install").Parse(string(templateContent))
	if err != nil {
		log.Printf("解析安装脚本模板失败: %v", err)
		c.String(http.StatusInternalServerError, "解析安装脚本模板失败")
		return
	}

	// 获取配置
	config := config.GetConfig()
	
	// 准备模板数据
	data := struct {
		PanelURL      string
		GithubRepo    string
		GithubVersion string
	}{
		PanelURL:      config.PanelURL,
		GithubRepo:    config.GithubRepo,
		GithubVersion: config.GithubVersion,
	}

	// 执行模板
	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		log.Printf("执行安装脚本模板失败: %v", err)
		c.String(http.StatusInternalServerError, "执行安装脚本模板失败")
		return
	}
}

// 生成节点安装命令
func GenerateInstallCommandHandler(c *gin.Context) {
	nodeID := c.Param("id")
	
	// 检查节点是否存在
	node, err := models.GetNode(nodeID)
	if err != nil {
		ErrorResponse(c, 404, fmt.Sprintf("节点不存在: %s", nodeID))
		return
	}

	// 获取面板URL
	panelURL := config.GetConfig().PanelURL
	if panelURL == "" {
		// 如果配置中没有设置面板URL，则使用请求中的Host
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		panelURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 生成节点密钥
	nodeKey := fmt.Sprintf("nk_%s_%d", node.ID, time.Now().Unix())

	// 保存节点密钥（实际项目中应该加密存储）
	node.SecretKey = nodeKey
	if err := models.SaveNode(node); err != nil {
		APIError(c, err)
		return
	}

	// 生成安装命令
	installCommand := fmt.Sprintf("curl -L %s/api/install.sh | bash -s -- %s \"%s\"", panelURL, nodeKey, node.Name)

	SuccessResponse(c, gin.H{
		"command": installCommand,
		"node_key": nodeKey,
		"panel_url": panelURL,
	})
}

// 节点程序下载处理函数
func DownloadNodeHandler(c *gin.Context) {
	arch := c.Param("arch") // 例如：node-amd64, node-arm64, node-arm

	// 验证节点密钥
	nodeKey := c.Query("key")
	if nodeKey == "" {
		c.String(http.StatusUnauthorized, "未提供节点密钥")
		return
	}

	// 验证节点密钥（实际项目中应该验证密钥是否有效）
	// 这里简化处理，只检查密钥格式
	if len(nodeKey) < 10 || nodeKey[:3] != "nk_" {
		c.String(http.StatusUnauthorized, "无效的节点密钥")
		return
	}

	// 根据架构选择节点程序文件
	var nodeBinary string
	switch arch {
	case "node-amd64":
		nodeBinary = "./bin/node-amd64"
	case "node-arm64":
		nodeBinary = "./bin/node-arm64"
	case "node-arm":
		nodeBinary = "./bin/node-arm"
	default:
		c.String(http.StatusBadRequest, "不支持的架构")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(nodeBinary); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "节点程序不存在")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(nodeBinary)))
	
	// 发送文件
	c.File(nodeBinary)
} 