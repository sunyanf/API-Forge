package service

import (
	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/model"
)

// GetUserByID fetches a user by ID
func GetUserByID(id uint) (*model.User, error) {
	return dao.GetUserByID(id)
}
