package models

import (
	"time"
)

// SpeedTestStatus 表示测速状态枚举
type SpeedTestStatus string

const (
	SpeedTestStatusPending   SpeedTestStatus = "pending"   // 等待中
	SpeedTestStatusRunning   SpeedTestStatus = "running"   // 运行中
	SpeedTestStatusCompleted SpeedTestStatus = "completed" // 已完成
	SpeedTestStatusFailed    SpeedTestStatus = "failed"    // 失败
	SpeedTestStatusTimeout   SpeedTestStatus = "timeout"   // 超时
)

// SpeedTestType 表示测速类型枚举
type SpeedTestType string

const (
	SpeedTestTypeDownload SpeedTestType = "download" // 下载测速
	SpeedTestTypeUpload   SpeedTestType = "upload"   // 上传测速
	SpeedTestTypePing     SpeedTestType = "ping"     // Ping测试
	SpeedTestTypeFull     SpeedTestType = "full"     // 全面测试
)

// SpeedTestResult 表示一次测速结果
type SpeedTestResult struct {
	ID            string         `json:"id"`             // 测试ID
	SourceNodeID  string         `json:"source_node_id"` // 源节点ID
	TargetNodeID  string         `json:"target_node_id"` // 目标节点ID
	Type          SpeedTestType  `json:"type"`           // 测试类型
	Status        SpeedTestStatus `json:"status"`        // 测试状态
	StartTime     time.Time      `json:"start_time"`     // 开始时间
	EndTime       time.Time      `json:"end_time"`       // 结束时间
	Duration      int64          `json:"duration"`       // 持续时间（毫秒）
	
	// 测速结果
	DownloadSpeed float64        `json:"download_speed"` // 下载速度（Mbps）
	UploadSpeed   float64        `json:"upload_speed"`   // 上传速度（Mbps）
	Ping          float64        `json:"ping"`           // Ping延迟（毫秒）
	Jitter        float64        `json:"jitter"`         // 抖动（毫秒）
	PacketLoss    float64        `json:"packet_loss"`    // 丢包率（百分比）
	
	// 错误信息
	ErrorMessage  string         `json:"error_message"`  // 错误信息
}

// SpeedTestRequest 表示测速请求
type SpeedTestRequest struct {
	SourceNodeID string        `json:"source_node_id"` // 源节点ID
	TargetNodeID string        `json:"target_node_id"` // 目标节点ID
	Type         SpeedTestType `json:"type"`           // 测试类型
	Timeout      int           `json:"timeout"`        // 超时时间（秒）
}

// SpeedTestResponse 表示测速请求响应
type SpeedTestResponse struct {
	ID       string `json:"id"`       // 测试ID
	Message  string `json:"message"`  // 响应消息
	Accepted bool   `json:"accepted"` // 是否接受测试
}

// SpeedTestResultList 表示测速结果列表
type SpeedTestResultList struct {
	Results []SpeedTestResult `json:"results"`
	Total   int              `json:"total"`
} 