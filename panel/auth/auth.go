package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// JWT密钥
var jwtSecret []byte

// 初始化JWT密钥
func InitJWTSecret(secret string) {
	if secret == "" {
		// 如果未提供密钥，生成随机密钥
		jwtSecret = make([]byte, 32)
		if _, err := rand.Read(jwtSecret); err != nil {
			log.Fatalf("无法生成JWT密钥: %v", err)
		}
		log.Println("已生成随机JWT密钥")
	} else {
		jwtSecret = []byte(secret)
		log.Println("使用配置的JWT密钥")
	}
}

// JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// 生成JWT令牌
func GenerateToken(userID, username, role string) (string, error) {
	// 设置过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "node-speedtest-panel",
			Subject:   userID,
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// 验证JWT令牌
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("无效的令牌")
}

// 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 生成API密钥
func GenerateAPIKey(userID string) (string, error) {
	// 生成32字节的随机数据
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	// 使用Base64编码
	key := base64.StdEncoding.EncodeToString(b)
	
	// 添加前缀和用户ID
	return fmt.Sprintf("nsp_%s_%s", userID, key), nil
}

// 验证节点密钥
func ValidateNodeKey(key string, nodeID string) bool {
	// 实际应用中应该从数据库验证
	// 这里简化处理，检查密钥格式和前缀
	if len(key) < 10 || key[:3] != "sk_" {
		return false
	}
	
	// 检查节点ID是否匹配
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return false
	}
	
	return parts[1] == nodeID
} 