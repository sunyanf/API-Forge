package model

import "time"

// User 用户账户模型
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`                          // 用户ID（主键）
	Email        string     `gorm:"size:255;uniqueIndex;not null" json:"email"`    // 邮箱（唯一索引）
	PasswordHash string     `gorm:"size:255;not null" json:"-"`                    // 密码哈希（JSON中隐藏）
	Name         string     `gorm:"size:128" json:"name"`                          // 昵称
	Role         string     `gorm:"size:32;default:user" json:"role"`              // 角色：user/vip/admin
	ApiKey       *string    `gorm:"size:36;uniqueIndex" json:"api_key,omitempty"`  // API密钥（唯一索引）
	CreatedAt    time.Time  `json:"created_at"`                                    // 注册时间
	UpdatedAt    time.Time  `json:"updated_at"`                                    // 更新时间
	LastLogin    *time.Time `json:"last_login"`                                    // 最后登录时间
}
