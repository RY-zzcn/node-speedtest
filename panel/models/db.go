package models

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

// 初始化数据库
func InitDB(dbPath string) error {
	var err error
	once.Do(func() {
		log.Printf("初始化数据库: %s", dbPath)
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Printf("打开数据库失败: %v", err)
			return
		}

		// 设置连接池参数
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)

		// 测试连接
		err = db.Ping()
		if err != nil {
			log.Printf("数据库连接测试失败: %v", err)
			return
		}

		// 创建表
		err = createTables()
		if err != nil {
			log.Printf("创建表失败: %v", err)
			return
		}
	})
	return err
}

// 创建数据库表
func createTables() error {
	// 创建节点表
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		ip TEXT NOT NULL,
		location TEXT,
		status TEXT NOT NULL,
		last_seen TIMESTAMP,
		created_at TIMESTAMP NOT NULL,
		description TEXT,
		tags TEXT,
		cpu REAL,
		memory REAL,
		disk REAL,
		uptime INTEGER,
		load1 REAL,
		load5 REAL,
		load15 REAL,
		network_rx INTEGER,
		network_tx INTEGER,
		version TEXT,
		secret_key TEXT
	)`)
	if err != nil {
		return fmt.Errorf("创建节点表失败: %v", err)
	}

	// 创建测速结果表
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS speedtest_results (
		id TEXT PRIMARY KEY,
		source_node_id TEXT NOT NULL,
		target_node_id TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP,
		duration INTEGER,
		download_speed REAL,
		upload_speed REAL,
		ping REAL,
		jitter REAL,
		packet_loss REAL,
		error_message TEXT,
		FOREIGN KEY (source_node_id) REFERENCES nodes (id),
		FOREIGN KEY (target_node_id) REFERENCES nodes (id)
	)`)
	if err != nil {
		return fmt.Errorf("创建测速结果表失败: %v", err)
	}

	// 创建用户表
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email TEXT,
		role TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		last_login TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	// 创建系统设置表
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("创建系统设置表失败: %v", err)
	}

	// 检查是否存在默认管理员用户，如果不存在则创建
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询管理员用户失败: %v", err)
	}

	if count == 0 {
		// 创建默认管理员用户 (用户名: admin, 密码: admin)
		_, err = db.Exec(`
		INSERT INTO users (id, username, password_hash, role, created_at)
		VALUES (?, ?, ?, ?, ?)`,
			"admin-"+generateID(),
			"admin",
			hashPassword("admin"), // 实际应用中应该使用更安全的密码哈希函数
			"admin",
			time.Now())
		if err != nil {
			return fmt.Errorf("创建默认管理员用户失败: %v", err)
		}
		log.Println("创建默认管理员用户: admin/admin")
	}

	return nil
}

// 保存节点
func SaveNode(node *Node) error {
	if node.ID == "" {
		node.ID = generateID()
	}
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now()
	}

	// 将标签数组转换为逗号分隔的字符串
	tags := joinTags(node.Tags)

	_, err := db.Exec(`
	INSERT OR REPLACE INTO nodes (
		id, name, ip, location, status, last_seen, created_at, description, tags,
		cpu, memory, disk, uptime, load1, load5, load15, network_rx, network_tx, version, secret_key
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		node.ID, node.Name, node.IP, node.Location, node.Status, node.LastSeen, node.CreatedAt,
		node.Description, tags, node.CPU, node.Memory, node.Disk, node.Uptime,
		node.Load[0], node.Load[1], node.Load[2], node.NetworkRx, node.NetworkTx, node.Version, node.SecretKey)

	return err
}

// 获取节点
func GetNode(id string) (*Node, error) {
	var node Node
	var tags string
	var load1, load5, load15 float64

	err := db.QueryRow(`
	SELECT id, name, ip, location, status, last_seen, created_at, description, tags,
		cpu, memory, disk, uptime, load1, load5, load15, network_rx, network_tx, version, secret_key
	FROM nodes WHERE id = ?`, id).Scan(
		&node.ID, &node.Name, &node.IP, &node.Location, &node.Status, &node.LastSeen, &node.CreatedAt,
		&node.Description, &tags, &node.CPU, &node.Memory, &node.Disk, &node.Uptime,
		&load1, &load5, &load15, &node.NetworkRx, &node.NetworkTx, &node.Version, &node.SecretKey)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("节点不存在: %s", id)
		}
		return nil, err
	}

	// 解析标签
	node.Tags = splitTags(tags)
	node.Load = [3]float64{load1, load5, load15}

	return &node, nil
}

// 获取所有节点
func GetAllNodes() ([]Node, error) {
	rows, err := db.Query(`
	SELECT id, name, ip, location, status, last_seen, created_at, description, tags,
		cpu, memory, disk, uptime, load1, load5, load15, network_rx, network_tx, version, secret_key
	FROM nodes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var tags string
		var load1, load5, load15 float64

		err := rows.Scan(
			&node.ID, &node.Name, &node.IP, &node.Location, &node.Status, &node.LastSeen, &node.CreatedAt,
			&node.Description, &tags, &node.CPU, &node.Memory, &node.Disk, &node.Uptime,
			&load1, &load5, &load15, &node.NetworkRx, &node.NetworkTx, &node.Version, &node.SecretKey)
		if err != nil {
			return nil, err
		}

		// 解析标签
		node.Tags = splitTags(tags)
		node.Load = [3]float64{load1, load5, load15}

		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

// 删除节点
func DeleteNode(id string) error {
	_, err := db.Exec("DELETE FROM nodes WHERE id = ?", id)
	return err
}

// 更新节点状态
func UpdateNodeStatus(id string, status NodeStatus) error {
	_, err := db.Exec("UPDATE nodes SET status = ?, last_seen = ? WHERE id = ?",
		status, time.Now(), id)
	return err
}

// 更新节点心跳
func UpdateNodeHeartbeat(heartbeat *NodeHeartbeat) error {
	_, err := db.Exec(`
	UPDATE nodes SET 
		last_seen = ?,
		status = ?,
		cpu = ?,
		memory = ?,
		disk = ?,
		uptime = ?,
		load1 = ?,
		load5 = ?,
		load15 = ?,
		network_rx = ?,
		network_tx = ?
	WHERE id = ?`,
		heartbeat.Timestamp,
		NodeStatusOnline,
		heartbeat.CPU,
		heartbeat.Memory,
		heartbeat.Disk,
		heartbeat.Uptime,
		heartbeat.Load[0],
		heartbeat.Load[1],
		heartbeat.Load[2],
		heartbeat.NetworkRx,
		heartbeat.NetworkTx,
		heartbeat.ID)
	return err
}

// 保存测速结果
func SaveSpeedTestResult(result *SpeedTestResult) error {
	if result.ID == "" {
		result.ID = generateID()
	}

	_, err := db.Exec(`
	INSERT OR REPLACE INTO speedtest_results (
		id, source_node_id, target_node_id, type, status, start_time, end_time,
		duration, download_speed, upload_speed, ping, jitter, packet_loss, error_message
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID, result.SourceNodeID, result.TargetNodeID, result.Type, result.Status,
		result.StartTime, result.EndTime, result.Duration, result.DownloadSpeed,
		result.UploadSpeed, result.Ping, result.Jitter, result.PacketLoss, result.ErrorMessage)

	return err
}

// 获取测速结果
func GetSpeedTestResult(id string) (*SpeedTestResult, error) {
	var result SpeedTestResult

	err := db.QueryRow(`
	SELECT id, source_node_id, target_node_id, type, status, start_time, end_time,
		duration, download_speed, upload_speed, ping, jitter, packet_loss, error_message
	FROM speedtest_results WHERE id = ?`, id).Scan(
		&result.ID, &result.SourceNodeID, &result.TargetNodeID, &result.Type, &result.Status,
		&result.StartTime, &result.EndTime, &result.Duration, &result.DownloadSpeed,
		&result.UploadSpeed, &result.Ping, &result.Jitter, &result.PacketLoss, &result.ErrorMessage)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("测速结果不存在: %s", id)
		}
		return nil, err
	}

	return &result, nil
}

// 获取所有测速结果
func GetAllSpeedTestResults() ([]SpeedTestResult, error) {
	rows, err := db.Query(`
	SELECT id, source_node_id, target_node_id, type, status, start_time, end_time,
		duration, download_speed, upload_speed, ping, jitter, packet_loss, error_message
	FROM speedtest_results ORDER BY start_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SpeedTestResult
	for rows.Next() {
		var result SpeedTestResult

		err := rows.Scan(
			&result.ID, &result.SourceNodeID, &result.TargetNodeID, &result.Type, &result.Status,
			&result.StartTime, &result.EndTime, &result.Duration, &result.DownloadSpeed,
			&result.UploadSpeed, &result.Ping, &result.Jitter, &result.PacketLoss, &result.ErrorMessage)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// 获取节点的测速结果
func GetNodeSpeedTestResults(nodeID string) ([]SpeedTestResult, error) {
	rows, err := db.Query(`
	SELECT id, source_node_id, target_node_id, type, status, start_time, end_time,
		duration, download_speed, upload_speed, ping, jitter, packet_loss, error_message
	FROM speedtest_results 
	WHERE source_node_id = ? OR target_node_id = ?
	ORDER BY start_time DESC`, nodeID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SpeedTestResult
	for rows.Next() {
		var result SpeedTestResult

		err := rows.Scan(
			&result.ID, &result.SourceNodeID, &result.TargetNodeID, &result.Type, &result.Status,
			&result.StartTime, &result.EndTime, &result.Duration, &result.DownloadSpeed,
			&result.UploadSpeed, &result.Ping, &result.Jitter, &result.PacketLoss, &result.ErrorMessage)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// 验证用户登录
func ValidateUser(username, password string) (bool, string, error) {
	var id string
	var passwordHash string

	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&id, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil // 用户不存在
		}
		return false, "", err
	}

	// 验证密码
	if verifyPassword(password, passwordHash) {
		// 更新最后登录时间
		_, err = db.Exec("UPDATE users SET last_login = ? WHERE id = ?", time.Now(), id)
		if err != nil {
			log.Printf("更新用户最后登录时间失败: %v", err)
		}
		return true, id, nil
	}

	return false, "", nil
}

// 获取设置值
func GetSetting(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // 设置不存在
		}
		return "", err
	}
	return value, nil
}

// 保存设置值
func SaveSetting(key, value string) error {
	_, err := db.Exec(`
	INSERT OR REPLACE INTO settings (key, value, updated_at)
	VALUES (?, ?, ?)`, key, value, time.Now())
	return err
}

// 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// 标签处理函数
func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := tags[0]
	for i := 1; i < len(tags); i++ {
		result += "," + tags[i]
	}
	return result
}

func splitTags(tags string) []string {
	if tags == "" {
		return []string{}
	}
	return splitString(tags, ",")
}

// 分割字符串
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return splitStringWithSep(s, sep)
}

// 实际分割字符串的函数
func splitStringWithSep(s, sep string) []string {
	// 这里使用简单的字符串分割
	// 在实际应用中，可能需要考虑更复杂的情况，如引号内的逗号等
	var result []string
	var current string
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// 密码哈希函数（简化版，实际应用中应使用更安全的方法如bcrypt）
func hashPassword(password string) string {
	// 这里简化处理，实际应用中应该使用bcrypt等安全哈希函数
	return fmt.Sprintf("hashed_%s", password)
}

// 验证密码
func verifyPassword(password, hash string) bool {
	// 这里简化处理，实际应用中应该使用bcrypt等安全哈希函数
	return hash == fmt.Sprintf("hashed_%s", password)
} 