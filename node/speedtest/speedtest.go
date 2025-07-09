package speedtest

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

// SpeedTestType 表示测速类型
type SpeedTestType string

const (
	TypeDownload SpeedTestType = "download" // 下载测速
	TypeUpload   SpeedTestType = "upload"   // 上传测速
	TypePing     SpeedTestType = "ping"     // Ping测试
	TypeFull     SpeedTestType = "full"     // 全面测试
)

// SpeedTestStatus 表示测速状态
type SpeedTestStatus string

const (
	StatusPending   SpeedTestStatus = "pending"   // 等待中
	StatusRunning   SpeedTestStatus = "running"   // 运行中
	StatusCompleted SpeedTestStatus = "completed" // 已完成
	StatusFailed    SpeedTestStatus = "failed"    // 失败
	StatusTimeout   SpeedTestStatus = "timeout"   // 超时
)

// SpeedTestResult 表示测速结果
type SpeedTestResult struct {
	ID            string          `json:"id"`             // 测试ID
	SourceNodeID  string          `json:"source_node_id"` // 源节点ID
	TargetNodeID  string          `json:"target_node_id"` // 目标节点ID
	Type          SpeedTestType   `json:"type"`           // 测试类型
	Status        SpeedTestStatus `json:"status"`         // 测试状态
	DownloadSpeed float64         `json:"download_speed"` // 下载速度（Mbps）
	UploadSpeed   float64         `json:"upload_speed"`   // 上传速度（Mbps）
	Ping          float64         `json:"ping"`           // Ping延迟（毫秒）
	Jitter        float64         `json:"jitter"`         // 抖动（毫秒）
	PacketLoss    float64         `json:"packet_loss"`    // 丢包率（百分比）
	StartTime     time.Time       `json:"start_time"`     // 开始时间
	EndTime       time.Time       `json:"end_time"`       // 结束时间
	Duration      int64           `json:"duration"`       // 持续时间（毫秒）
	Error         string          `json:"error_message"`  // 错误信息
}

// SpeedTestRequest 表示测速请求
type SpeedTestRequest struct {
	ID           string        `json:"id"`            // 测试ID
	SourceNodeID string        `json:"source_node_id"`// 源节点ID
	TargetNodeID string        `json:"target_node_id"`// 目标节点ID
	TargetURL    string        `json:"target_url"`    // 目标URL
	Type         SpeedTestType `json:"type"`          // 测试类型
	Timeout      int           `json:"timeout"`       // 超时时间（秒）
	Threads      int           `json:"threads"`       // 线程数
}

// 测速管理器
type SpeedTestManager struct {
	activeTests map[string]*SpeedTestResult
	mutex       sync.RWMutex
	panelURL    string
	nodeID      string
	nodeKey     string
	httpClient  *http.Client
}

// 创建新的测速管理器
func NewSpeedTestManager(panelURL, nodeID, nodeKey string) *SpeedTestManager {
	return &SpeedTestManager{
		activeTests: make(map[string]*SpeedTestResult),
		panelURL:    panelURL,
		nodeID:      nodeID,
		nodeKey:     nodeKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// 启动测速
func (m *SpeedTestManager) StartTest(req SpeedTestRequest) (*SpeedTestResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 创建测试结果
	result := &SpeedTestResult{
		ID:           req.ID,
		SourceNodeID: req.SourceNodeID,
		TargetNodeID: req.TargetNodeID,
		Type:         req.Type,
		Status:       StatusRunning,
		StartTime:    time.Now(),
	}
	
	// 存储活跃测试
	m.activeTests[req.ID] = result
	
	// 启动测速协程
	go func() {
		var err error
		
		// 设置测试超时
		timeout := time.Duration(req.Timeout) * time.Second
		if timeout == 0 {
			timeout = 120 * time.Second // 默认120秒
		}
		
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		
		// 创建完成通道
		done := make(chan struct{})
		
		// 启动测试协程
		go func() {
			defer close(done)
			
			// 根据测试类型执行不同的测试
			switch req.Type {
			case TypeDownload:
				err = m.runDownloadTest(req, result)
			case TypeUpload:
				err = m.runUploadTest(req, result)
			case TypePing:
				err = m.runPingTest(req, result)
			case TypeFull:
				err = m.runFullTest(req, result)
			default:
				err = errors.New("未知的测试类型")
			}
		}()
		
		// 等待测试完成或超时
		select {
		case <-done:
			// 测试正常完成
		case <-ctx.Done():
			// 测试超时
			err = errors.New("测试超时")
			result.Status = StatusTimeout
		}
		
		// 完成测试
		m.mutex.Lock()
		defer m.mutex.Unlock()
		
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
		
		if err != nil {
			result.Error = err.Error()
			if result.Status != StatusTimeout {
				result.Status = StatusFailed
			}
			log.Printf("测速失败: %v", err)
		} else {
			result.Status = StatusCompleted
			log.Printf("测速完成: %+v", *result)
		}
		
		// 上报结果到面板
		go m.reportTestResult(*result)
		
		// 一段时间后清理测试结果
		go func() {
			time.Sleep(5 * time.Minute)
			m.mutex.Lock()
			defer m.mutex.Unlock()
			delete(m.activeTests, req.ID)
		}()
	}()
	
	return result, nil
}

// 获取测试结果
func (m *SpeedTestManager) GetTestResult(testID string) (*SpeedTestResult, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result, exists := m.activeTests[testID]
	return result, exists
}

// 获取所有活跃测试
func (m *SpeedTestManager) GetAllTests() []SpeedTestResult {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	results := make([]SpeedTestResult, 0, len(m.activeTests))
	for _, test := range m.activeTests {
		results = append(results, *test)
	}
	
	return results
}

// 执行下载测速
func (m *SpeedTestManager) runDownloadTest(req SpeedTestRequest, result *SpeedTestResult) error {
	log.Printf("开始下载测速: %s -> %s", req.SourceNodeID, req.TargetNodeID)
	
	// 确定目标URL
	targetURL := req.TargetURL
	if targetURL == "" {
		// 如果没有提供URL，使用默认测试URL
		targetURL = fmt.Sprintf("%s/speedtest/download?size=100", m.panelURL)
	}
	
	// 确定线程数
	threads := req.Threads
	if threads <= 0 {
		threads = 4 // 默认4个线程
	}
	
	// 准备测试
	var totalBytes int64
	var totalDuration time.Duration
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 启动多个线程进行下载测试
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			
			// 创建HTTP请求
			req, err := http.NewRequest("GET", targetURL, nil)
			if err != nil {
				log.Printf("创建HTTP请求失败: %v", err)
				return
			}
			
			// 设置请求头
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("User-Agent", "NodeSpeedTest/1.0")
			
			// 执行请求
			startTime := time.Now()
			resp, err := m.httpClient.Do(req)
			if err != nil {
				log.Printf("执行HTTP请求失败: %v", err)
				return
			}
			defer resp.Body.Close()
			
			// 读取响应内容
			n, err := io.Copy(ioutil.Discard, resp.Body)
			if err != nil {
				log.Printf("读取响应内容失败: %v", err)
				return
			}
			
			duration := time.Since(startTime)
			
			// 更新统计信息
			mu.Lock()
			totalBytes += n
			totalDuration += duration
			mu.Unlock()
			
			log.Printf("下载测试线程 %d 完成: %d 字节, 耗时 %v", threadID, n, duration)
		}(i)
	}
	
	// 等待所有线程完成
	wg.Wait()
	
	// 计算下载速度（Mbps）
	if totalDuration.Seconds() > 0 {
		// 计算平均下载速度（Mbps）
		// 公式: (总字节数 * 8) / (总时间 * 1000000) = Mbps
		result.DownloadSpeed = float64(totalBytes) * 8 / totalDuration.Seconds() / 1000000
		log.Printf("下载测速结果: %.2f Mbps", result.DownloadSpeed)
	} else {
		return errors.New("下载测试时间过短")
	}
	
	return nil
}

// 执行上传测速
func (m *SpeedTestManager) runUploadTest(req SpeedTestRequest, result *SpeedTestResult) error {
	log.Printf("开始上传测速: %s -> %s", req.SourceNodeID, req.TargetNodeID)
	
	// 确定目标URL
	targetURL := req.TargetURL
	if targetURL == "" {
		// 如果没有提供URL，使用默认测试URL
		targetURL = fmt.Sprintf("%s/speedtest/upload", m.panelURL)
	}
	
	// 确定线程数
	threads := req.Threads
	if threads <= 0 {
		threads = 2 // 默认2个线程
	}
	
	// 准备测试
	var totalBytes int64
	var totalDuration time.Duration
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 生成随机数据（10MB）
	dataSize := 10 * 1024 * 1024 // 10MB
	data := make([]byte, dataSize)
	if _, err := rand.Read(data); err != nil {
		return fmt.Errorf("生成随机数据失败: %v", err)
	}
	
	// 启动多个线程进行上传测试
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			
			// 创建HTTP请求
			req, err := http.NewRequest("POST", targetURL, bytes.NewReader(data))
			if err != nil {
				log.Printf("创建HTTP请求失败: %v", err)
				return
			}
			
			// 设置请求头
			req.Header.Set("Content-Type", "application/octet-stream")
			req.Header.Set("User-Agent", "NodeSpeedTest/1.0")
			
			// 执行请求
			startTime := time.Now()
			resp, err := m.httpClient.Do(req)
			if err != nil {
				log.Printf("执行HTTP请求失败: %v", err)
				return
			}
			defer resp.Body.Close()
			
			// 读取响应内容
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("读取响应内容失败: %v", err)
				return
			}
			
			duration := time.Since(startTime)
			
			// 更新统计信息
			mu.Lock()
			totalBytes += int64(dataSize)
			totalDuration += duration
			mu.Unlock()
			
			log.Printf("上传测试线程 %d 完成: %d 字节, 耗时 %v", threadID, dataSize, duration)
		}(i)
	}
	
	// 等待所有线程完成
	wg.Wait()
	
	// 计算上传速度（Mbps）
	if totalDuration.Seconds() > 0 {
		// 计算平均上传速度（Mbps）
		// 公式: (总字节数 * 8) / (总时间 * 1000000) = Mbps
		result.UploadSpeed = float64(totalBytes) * 8 / totalDuration.Seconds() / 1000000
		log.Printf("上传测速结果: %.2f Mbps", result.UploadSpeed)
	} else {
		return errors.New("上传测试时间过短")
	}
	
	return nil
}

// 执行Ping测试
func (m *SpeedTestManager) runPingTest(req SpeedTestRequest, result *SpeedTestResult) error {
	log.Printf("开始Ping测试: %s -> %s", req.SourceNodeID, req.TargetNodeID)
	
	// 确定目标主机
	host := req.TargetURL
	if host == "" {
		// 如果没有提供URL，尝试解析目标节点的IP
		// 实际应用中应该从面板获取节点IP
		host = fmt.Sprintf("%s/speedtest/ping", m.panelURL)
	}
	
	// 准备测试
	count := 10 // 默认ping 10次
	var durations []time.Duration
	var packetLoss int
	
	// 执行HTTP ping测试
	for i := 0; i < count; i++ {
		// 创建HTTP请求
		req, err := http.NewRequest("GET", host, nil)
		if err != nil {
			log.Printf("创建HTTP请求失败: %v", err)
			packetLoss++
			continue
		}
		
		// 设置请求头
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("User-Agent", "NodeSpeedTest/1.0")
		
		// 执行请求
		startTime := time.Now()
		resp, err := m.httpClient.Do(req)
		if err != nil {
			log.Printf("执行HTTP请求失败: %v", err)
			packetLoss++
			continue
		}
		
		// 读取响应内容
		_, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("读取响应内容失败: %v", err)
			packetLoss++
			continue
		}
		
		duration := time.Since(startTime)
		durations = append(durations, duration)
		
		// 等待一段时间再进行下一次测试
		time.Sleep(100 * time.Millisecond)
	}
	
	// 计算ping结果
	if len(durations) > 0 {
		// 计算平均延迟（毫秒）
		var totalDuration time.Duration
		for _, d := range durations {
			totalDuration += d
		}
		result.Ping = float64(totalDuration.Milliseconds()) / float64(len(durations))
		
		// 计算抖动（毫秒）
		if len(durations) > 1 {
			var jitterSum float64
			for i := 1; i < len(durations); i++ {
				jitter := math.Abs(float64(durations[i].Milliseconds() - durations[i-1].Milliseconds()))
				jitterSum += jitter
			}
			result.Jitter = jitterSum / float64(len(durations)-1)
		}
		
		// 计算丢包率（百分比）
		result.PacketLoss = float64(packetLoss) / float64(count) * 100
		
		log.Printf("Ping测试结果: %.2f ms, 抖动: %.2f ms, 丢包率: %.2f%%", 
			result.Ping, result.Jitter, result.PacketLoss)
	} else {
		return errors.New("Ping测试失败，无有效结果")
	}
	
	return nil
}

// 执行全面测试
func (m *SpeedTestManager) runFullTest(req SpeedTestRequest, result *SpeedTestResult) error {
	log.Printf("开始全面测速: %s -> %s", req.SourceNodeID, req.TargetNodeID)
	
	// 先执行Ping测试
	if err := m.runPingTest(req, result); err != nil {
		return fmt.Errorf("Ping测试失败: %v", err)
	}
	
	// 执行下载测速
	if err := m.runDownloadTest(req, result); err != nil {
		return fmt.Errorf("下载测速失败: %v", err)
	}
	
	// 执行上传测速
	if err := m.runUploadTest(req, result); err != nil {
		return fmt.Errorf("上传测速失败: %v", err)
	}
	
	return nil
}

// 上报测试结果到面板
func (m *SpeedTestManager) reportTestResult(result SpeedTestResult) {
	// 准备请求体
	body, err := json.Marshal(result)
	if err != nil {
		log.Printf("序列化测试结果失败: %v", err)
		return
	}
	
	// 创建HTTP请求
	url := fmt.Sprintf("%s/api/node/speedtest/result", m.panelURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("创建HTTP请求失败: %v", err)
		return
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Node-Key", m.nodeKey)
	
	// 执行请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Printf("上报测试结果失败: %v", err)
		return
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		log.Printf("上报测试结果失败，状态码: %d", resp.StatusCode)
		return
	}
	
	log.Printf("成功上报测试结果: %s", result.ID)
} 