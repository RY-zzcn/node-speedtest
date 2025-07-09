package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// Config 表示节点的配置结构
type Config struct {
	// 基本配置
	ListenPort   string `json:"listen_port"`    // 监听端口
	PanelURL     string `json:"panel_url"`      // 面板URL
	NodeKey      string `json:"node_key"`       // 节点密钥
	NodeName     string `json:"node_name"`      // 节点名称
	LogPath      string `json:"log_path"`       // 日志路径
	DataDir      string `json:"data_dir"`       // 数据目录
	
	// 心跳配置
	HeartbeatInterval int `json:"heartbeat_interval"` // 心跳间隔（秒）
	
	// 测速配置
	SpeedtestTimeout int `json:"speedtest_timeout"`   // 测速超时（秒）
	DownloadThreads  int `json:"download_threads"`    // 下载测试线程数
	UploadThreads    int `json:"upload_threads"`      // 上传测试线程数
	PingCount        int `json:"ping_count"`          // Ping测试次数
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
			ListenPort:        "8081",
			PanelURL:          "",
			NodeKey:           "",
			NodeName:          "",
			LogPath:           "./node.log",
			DataDir:           "./data",
			HeartbeatInterval: 30,
			SpeedtestTimeout:  120,
			DownloadThreads:   4,
			UploadThreads:     2,
			PingCount:         10,
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

// UpdateConfig 更新配置
func UpdateConfig(newConfig Config) {
	mu.Lock()
	defer mu.Unlock()
	
	// 更新配置
	config.ListenPort = newConfig.ListenPort
	config.PanelURL = newConfig.PanelURL
	config.NodeKey = newConfig.NodeKey
	config.NodeName = newConfig.NodeName
	config.LogPath = newConfig.LogPath
	config.DataDir = newConfig.DataDir
	config.HeartbeatInterval = newConfig.HeartbeatInterval
	config.SpeedtestTimeout = newConfig.SpeedtestTimeout
	config.DownloadThreads = newConfig.DownloadThreads
	config.UploadThreads = newConfig.UploadThreads
	config.PingCount = newConfig.PingCount
	
	// 保存到文件
	saveConfig()
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