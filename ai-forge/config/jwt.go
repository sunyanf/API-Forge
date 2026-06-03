package config

import (
	"os"
	"strconv"
)

// JWTSecret returns the JWT signing secret from env or a default for dev
func JWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "dev-secret"
}

// JWTExpiryMinutes returns token expiry in minutes (default 60)
func JWTExpiryMinutes() int {
	if s := os.Getenv("JWT_EXPIRY_MINUTES"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return 60
}
