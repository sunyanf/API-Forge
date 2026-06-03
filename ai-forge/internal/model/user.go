package model

import (
	"time"
)

// User is the GORM model for users table
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Name         string `gorm:"size:128"`
	Role         string `gorm:"size:32;default:'user'"`
	ApiKey       *string `gorm:"size:36;uniqueIndex" json:"api_key,omitempty"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    *time.Time
}
