package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sunyanf/ai-forge/config"
	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/model"
)

// AuthenticateUser 验证用户凭据，成功则返回用户对象
func AuthenticateUser(email, password string) (*model.User, error) {
	u, err := dao.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if u == nil || u.ID == 0 {
		return nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return u, nil
}

// GenerateJWT 为指定用户生成签名的 JWT Token，包含用户 ID、邮箱和角色
func GenerateJWT(u *model.User) (string, error) {
	secret := config.JWTSecret()
	exp := time.Now().Add(time.Duration(config.JWTExpiryMinutes()) * time.Minute)
	role := u.Role
	if role == "" {
		role = "user"
	}
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", u.ID),
		"email": u.Email,
		"role":  role,
		"exp":   exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
