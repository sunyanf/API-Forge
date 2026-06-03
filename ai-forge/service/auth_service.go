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

// AuthenticateUser verifies credentials and returns the user if ok
func AuthenticateUser(email, password string) (*model.User, error) {
	u, err := dao.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if u == nil || u.ID == 0 {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// GenerateJWT creates a signed JWT for the given user
func GenerateJWT(u *model.User) (string, error) {
	secret := config.JWTSecret()
	exp := time.Now().Add(time.Duration(config.JWTExpiryMinutes()) * time.Minute)
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", u.ID),
		"email": u.Email,
		"exp":   exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
