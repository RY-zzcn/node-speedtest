package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 节点认证结构体
type NodeAuth struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// API处理器
type APIHandler struct {
	config     *Config
	installSh  []byte
	nodeFiles  map[string]string
}

// 配置结构体
type Config struct {
	PanelURL    string `json:"panel_url"`
	SecretKey   string `json:"secret_key"`
	GithubRepo  string `json:"github_repo"`
	GithubVersion string `json:"github_version"`
}

// 创建新的API处理器
func NewAPIHandler(config *Config) (*APIHandler, error) {
	// 读取安装脚本
	installSh, err := os.ReadFile("api/install.sh")
	if err != nil {
		return nil, fmt.Errorf("读取安装脚本失败: %v", err)
	}

	// 节点程序文件路径映射
	nodeFiles := map[string]string{
		"amd64": "bin/node-amd64",
		"arm64": "bin/node-arm64",
		"arm":   "bin/node-arm",
	}

	return &APIHandler{
		config:    config,
		installSh: installSh,
		nodeFiles: nodeFiles,
	}, nil
}

// 处理API请求
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// 处理安装脚本请求
	if path == "/api/install.sh" {
		h.handleInstallScript(w, r)
		return
	}
	
	// 处理节点程序下载请求
	if strings.HasPrefix(path, "/api/download/node-") {
		h.handleNodeDownload(w, r)
		return
	}
	
	// 处理API ping请求
	if path == "/api/ping" {
		h.handlePing(w, r)
		return
	}
	
	// 其他API请求处理...
	
	// 默认返回404
	http.NotFound(w, r)
}

// 处理安装脚本请求
func (h *APIHandler) handleInstallScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(h.installSh)
}

// 处理节点程序下载请求
func (h *APIHandler) handleNodeDownload(w http.ResponseWriter, r *http.Request) {
	// 获取节点架构
	arch := strings.TrimPrefix(r.URL.Path, "/api/download/node-")
	
	// 获取节点密钥
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "缺少节点密钥", http.StatusBadRequest)
		return
	}
	
	// 验证节点密钥（实际应用中应该查询数据库验证）
	if !h.validateNodeKey(key) {
		http.Error(w, "无效的节点密钥", http.StatusUnauthorized)
		return
	}
	
	// 获取节点程序文件路径
	filePath, ok := h.nodeFiles[arch]
	if !ok {
		http.Error(w, "不支持的架构", http.StatusBadRequest)
		return
	}
	
	// 打开节点程序文件
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "节点程序文件不存在", http.StatusNotFound)
		return
	}
	defer file.Close()
	
	// 设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=node-%s", arch))
	
	// 发送文件
	io.Copy(w, file)
}

// 处理API ping请求
func (h *APIHandler) handlePing(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
		"version": h.config.GithubVersion,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 验证节点密钥
func (h *APIHandler) validateNodeKey(key string) bool {
	// 实际应用中应该查询数据库验证
	// 这里简单实现，只检查密钥长度和格式
	return len(key) == 32 && !strings.Contains(key, " ")
}

// 注册API路由
func RegisterAPIRoutes(mux *http.ServeMux, config *Config) error {
	handler, err := NewAPIHandler(config)
	if err != nil {
		return err
	}
	
	mux.Handle("/api/", handler)
	return nil
} 