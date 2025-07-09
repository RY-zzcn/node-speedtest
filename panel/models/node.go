package models

import (
	"time"
)

// NodeStatus 表示节点状态枚举
type NodeStatus string

const (
	NodeStatusOnline  NodeStatus = "online"  // 在线
	NodeStatusOffline NodeStatus = "offline" // 离线
	NodeStatusError   NodeStatus = "error"   // 错误
)

// Node 表示一个节点
type Node struct {
	ID          string     `json:"id"`           // 节点唯一标识
	Name        string     `json:"name"`         // 节点名称
	IP          string     `json:"ip"`           // 节点IP地址
	Location    string     `json:"location"`     // 节点地理位置
	Status      NodeStatus `json:"status"`       // 节点状态
	LastSeen    time.Time  `json:"last_seen"`    // 最后一次心跳时间
	CreatedAt   time.Time  `json:"created_at"`   // 创建时间
	Description string     `json:"description"`  // 节点描述
	Tags        []string   `json:"tags"`         // 节点标签
	
	// 系统信息
	CPU         float64    `json:"cpu"`          // CPU使用率
	Memory      float64    `json:"memory"`       // 内存使用率
	Disk        float64    `json:"disk"`         // 硬盘使用率
	Uptime      int64      `json:"uptime"`       // 运行时间（秒）
	Load        [3]float64 `json:"load"`         // 系统负载（1分钟、5分钟、15分钟）
	
	// 网络信息
	NetworkRx   int64      `json:"network_rx"`   // 网络接收字节数
	NetworkTx   int64      `json:"network_tx"`   // 网络发送字节数
	
	// 版本信息
	Version     string     `json:"version"`      // 节点客户端版本
	
	// 安全信息
	SecretKey   string     `json:"-"`           // 节点密钥（不输出到JSON）
}

// NodeList 表示节点列表
type NodeList struct {
	Nodes []Node `json:"nodes"`
	Total int    `json:"total"`
}

// NodeHeartbeat 表示节点心跳数据
type NodeHeartbeat struct {
	ID        string     `json:"id"`         // 节点ID
	Timestamp time.Time  `json:"timestamp"`  // 心跳时间
	CPU       float64    `json:"cpu"`        // CPU使用率
	Memory    float64    `json:"memory"`     // 内存使用率
	Disk      float64    `json:"disk"`       // 硬盘使用率
	Uptime    int64      `json:"uptime"`     // 运行时间
	Load      [3]float64 `json:"load"`       // 系统负载
	NetworkRx int64      `json:"network_rx"` // 网络接收
	NetworkTx int64      `json:"network_tx"` // 网络发送
}

// NodeRegisterRequest 表示节点注册请求
type NodeRegisterRequest struct {
	Name        string   `json:"name"`        // 节点名称
	IP          string   `json:"ip"`          // 节点IP
	Location    string   `json:"location"`    // 地理位置
	Description string   `json:"description"` // 描述
	Tags        []string `json:"tags"`        // 标签
	Version     string   `json:"version"`     // 版本
}

// NodeRegisterResponse 表示节点注册响应
type NodeRegisterResponse struct {
	ID        string `json:"id"`        // 分配的节点ID
	SecretKey string `json:"secretKey"` // 用于认证的密钥
} 