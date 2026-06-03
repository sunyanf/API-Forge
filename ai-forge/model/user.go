package model

import "time"

// User represents a user account
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Email        string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Name         string     `gorm:"size:128" json:"name"`
	Role         string     `gorm:"size:32;default:user" json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login"`
}
