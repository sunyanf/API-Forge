package dao

import (
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
)

// GetUserByID returns a user by primary key
func GetUserByID(id uint) (*model.User, error) {
	var u model.User
	if err := db.DB.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
