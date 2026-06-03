package dao

import (
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
)

func CreateUser(u *model.User) error {
	return db.DB.Create(u).Error
}

func GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := db.DB.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
