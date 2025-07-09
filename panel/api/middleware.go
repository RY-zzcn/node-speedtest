package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"../auth"
)

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是公开API
		if isPublicAPI(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 从Cookie中获取会话令牌
		token, err := c.Cookie("session_token")
		if err != nil {
			// 从请求头中获取令牌
			token = c.GetHeader("Authorization")
			// 移除Bearer前缀（如果有）
			token = strings.TrimPrefix(token, "Bearer ")
		}

		if token == "" {
			RequireLogin(c)
			return
		}

		// 验证令牌
		claims, err := auth.ValidateToken(token)
		if err != nil {
			log.Printf("无效的令牌: %v", err)
			RequireLogin(c)
			return
		}

		// 将用户信息存储在上下文中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// 节点认证中间件
func NodeAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取节点密钥
		nodeKey := c.GetHeader("Node-Key")
		if nodeKey == "" {
			ErrorResponse(c, 401, "缺少节点密钥")
			c.Abort()
			return
		}

		// 验证节点密钥格式
		if !strings.HasPrefix(nodeKey, "sk_") {
			ErrorResponse(c, 401, "无效的节点密钥格式")
			c.Abort()
			return
		}

		// 解析密钥中的节点ID
		parts := strings.Split(nodeKey, "_")
		if len(parts) < 3 {
			ErrorResponse(c, 401, "无效的节点密钥格式")
			c.Abort()
			return
		}

		nodeID := parts[1]

		// 验证节点密钥
		// TODO: 从数据库验证节点密钥
		if !auth.ValidateNodeKey(nodeKey, nodeID) {
			ErrorResponse(c, 401, "无效的节点密钥")
			c.Abort()
			return
		}

		// 将节点ID存储在上下文中
		c.Set("nodeID", nodeID)

		c.Next()
	}
}

// 管理员权限中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := c.Get("userRole")
		if !exists {
			RequireLogin(c)
			return
		}

		// 检查是否是管理员
		if userRole != "admin" {
			ErrorResponse(c, 403, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Node-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latency := endTime.Sub(startTime)

		// 请求方法
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		log.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s",
			endTime.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}

// 检查是否是公开API
func isPublicAPI(path string) bool {
	publicPaths := []string{
		"/api/login",
		"/api/register",
		"/api/node/register",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
} 