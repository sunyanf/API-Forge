// 文件名: user_dao_extra.go
// 作用: 用户数据访问层的补充查询函数

package dao

import (
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
)

// GetUserByID 根据主键 ID 查询用户
func GetUserByID(id uint) (*model.User, error) {
	var u model.User
	if err := db.DB.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByAPIKey 根据 API Key 查询用户
func GetUserByAPIKey(apiKey string) (*model.User, error) {
	var u model.User
	if err := db.DB.Where("api_key = ?", apiKey).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAllUsers 获取所有用户列表（管理员功能）
func GetAllUsers() ([]model.User, error) {
	var users []model.User
	if err := db.DB.Order("id desc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserRole 更新指定用户的角色
func UpdateUserRole(userID uint, role string) error {
	return db.DB.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error
}

// UpdateUserPassword 更新指定用户的密码哈希
func UpdateUserPassword(userID uint, passwordHash string) error {
	return db.DB.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}
