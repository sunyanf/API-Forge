package config

import (
	"os"
	"strconv"
)

// JWTSecret 返回 JWT 签名密钥（优先使用环境变量，开发环境使用默认值）
func JWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "dev-secret"
}

// JWTExpiryMinutes 返回 Token 过期时间（分钟，默认 60 分钟）
func JWTExpiryMinutes() int {
	if s := os.Getenv("JWT_EXPIRY_MINUTES"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return 60
}
