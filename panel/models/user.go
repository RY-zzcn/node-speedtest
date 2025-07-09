package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"../auth"
)

// 用户角色
type UserRole string

const (
	RoleAdmin  UserRole = "admin"  // 管理员
	RoleUser   UserRole = "user"   // 普通用户
	RoleViewer UserRole = "viewer" // 只读用户
)

// 用户模型
type User struct {
	ID           string    `json:"id"`            // 用户ID
	Username     string    `json:"username"`      // 用户名
	PasswordHash string    `json:"-"`             // 密码哈希（不返回给前端）
	Email        string    `json:"email"`         // 邮箱
	Role         string    `json:"role"`          // 角色
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	LastLogin    time.Time `json:"last_login"`    // 最后登录时间
	APIKey       string    `json:"-"`             // API密钥（不返回给前端）
}

// 保存用户
func SaveUser(user *User) error {
	if user.ID == "" {
		user.ID = generateID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := db.Exec(`
	INSERT OR REPLACE INTO users (
		id, username, password_hash, email, role, created_at, last_login
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.Email, user.Role,
		user.CreatedAt, user.LastLogin)

	return err
}

// 获取用户
func GetUser(id string) (*User, error) {
	var user User

	err := db.QueryRow(`
	SELECT id, username, password_hash, email, role, created_at, last_login
	FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
		&user.CreatedAt, &user.LastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在: %s", id)
		}
		return nil, err
	}

	return &user, nil
}

// 获取用户（按用户名）
func GetUserByUsername(username string) (*User, error) {
	var user User

	err := db.QueryRow(`
	SELECT id, username, password_hash, email, role, created_at, last_login
	FROM users WHERE username = ?`, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
		&user.CreatedAt, &user.LastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在: %s", username)
		}
		return nil, err
	}

	return &user, nil
}

// 获取所有用户
func GetAllUsers() ([]User, error) {
	rows, err := db.Query(`
	SELECT id, username, password_hash, email, role, created_at, last_login
	FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
			&user.CreatedAt, &user.LastLogin)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// 删除用户
func DeleteUser(id string) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// 更新用户最后登录时间
func UpdateUserLastLogin(id string) error {
	_, err := db.Exec("UPDATE users SET last_login = ? WHERE id = ?", time.Now(), id)
	return err
}

// 更新用户密码
func UpdateUserPassword(id string, newPassword string) error {
	// 哈希新密码
	passwordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, id)
	return err
}

// 检查用户名是否存在
func UserExists(username string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 验证用户登录
func ValidateUser(username, password string) (bool, string, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		if err.Error() == fmt.Sprintf("用户不存在: %s", username) {
			return false, "", nil // 用户不存在
		}
		return false, "", err
	}

	// 验证密码
	if auth.CheckPassword(password, user.PasswordHash) {
		// 更新最后登录时间
		if err := UpdateUserLastLogin(user.ID); err != nil {
			// 记录错误但不中断流程
			log.Printf("更新用户最后登录时间失败: %v", err)
		}
		return true, user.ID, nil
	}

	return false, "", nil
} 