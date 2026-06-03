package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/model"
)

// RegisterUser creates a new user with plaintext password (will be hashed)
func RegisterUser(email, password, name string) (*model.User, error) {
	// check existing
	if u, _ := dao.GetUserByEmail(email); u != nil && u.ID != 0 {
		return nil, errors.New("user already exists")
	}
	// hash password
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{Email: email, PasswordHash: string(h), Name: name}
	if err := dao.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}
