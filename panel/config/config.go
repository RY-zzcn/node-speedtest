package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// Config 表示面板的配置结构
type Config struct {
	// 基本配置
	ListenPort     string `json:"listen_port"`      // 监听端口
	DatabasePath   string `json:"database_path"`    // 数据库路径
	LogPath        string `json:"log_path"`         // 日志路径
	SecretKey      string `json:"secret_key"`       // 用于加密通信的密钥
	AdminUsername  string `json:"admin_username"`   // 管理员用户名
	AdminPassword  string `json:"admin_password"`   // 管理员密码（存储为哈希值）
	PanelURL       string `json:"panel_url"`        // 面板URL，用于节点连接
	
	// 节点配置
	NodeTimeout    int    `json:"node_timeout"`     // 节点超时时间（秒）
	NodeCheckInterval int  `json:"node_check_interval"` // 节点检查间隔（秒）
	
	// 测速配置
	SpeedtestTimeout int  `json:"speedtest_timeout"` // 测速超时时间（秒）
	MaxConcurrentTests int `json:"max_concurrent_tests"` // 最大并发测试数
	
	// GitHub配置
	GithubRepo    string `json:"github_repo"`     // GitHub仓库地址
	GithubVersion string `json:"github_version"`  // GitHub发布版本
}

var (
	config *Config
	once   sync.Once
	mu     sync.RWMutex
	configPath string
)

// SetConfigPath 设置配置文件路径
func SetConfigPath(path string) {
	configPath = path
}

// GetConfig 获取配置单例
func GetConfig() *Config {
	once.Do(func() {
		config = &Config{
			// 默认配置
			ListenPort:        "8080",
			DatabasePath:      "./data.db",
			LogPath:           "./panel.log",
			SecretKey:         "change_this_to_a_random_string",
			AdminUsername:     "admin",
			AdminPassword:     "admin", // 默认密码，应该在首次使用时要求更改
			NodeTimeout:       60,
			NodeCheckInterval: 30,
			SpeedtestTimeout:  120,
			MaxConcurrentTests: 3,
		}
		
		// 尝试从文件加载配置
		if configPath != "" {
			loadConfig()
		}
	})
	
	mu.RLock()
	defer mu.RUnlock()
	return config
}

// 从文件加载配置
func loadConfig() {
	mu.Lock()
	defer mu.Unlock()
	
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		saveConfig()
		return
	}
	
	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("读取配置文件失败: %v，将使用默认配置", err)
		return
	}
	
	// 解析配置
	if err := json.Unmarshal(data, config); err != nil {
		log.Printf("解析配置文件失败: %v，将使用默认配置", err)
		return
	}
	
	log.Printf("成功从 %s 加载配置", configPath)
}

// SaveConfig 保存配置到文件
func SaveConfig() error {
	mu.Lock()
	defer mu.Unlock()
	
	return saveConfig()
}

// 保存配置到文件（内部使用，已加锁）
func saveConfig() error {
	// 将配置序列化为JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("序列化配置失败: %v", err)
		return err
	}
	
	// 写入文件
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		log.Printf("写入配置文件失败: %v", err)
		return err
	}
	
	log.Printf("成功保存配置到 %s", configPath)
	return nil
} 