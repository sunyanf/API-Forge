package service

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/model"
)

// RegisterUser 注册新用户，密码会被哈希存储
func RegisterUser(email, password, name string) (*model.User, error) {
	// 检查是否已存在
	if u, _ := dao.GetUserByEmail(email); u != nil && u.ID != 0 {
		return nil, errors.New("该邮箱已被注册")
	}
	// 密码哈希
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	apiKey := uuid.NewString()
	u := &model.User{Email: email, PasswordHash: string(h), Name: name, ApiKey: &apiKey}
	if err := dao.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

// ChangePassword 修改指定用户的密码
// 参数 userID: 用户 ID
// 参数 newPassword: 新密码（明文，内部会哈希）
func ChangePassword(userID uint, newPassword string) error {
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return dao.UpdateUserPassword(userID, string(h))
}
