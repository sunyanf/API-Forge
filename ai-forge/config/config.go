package config

import (
	"os"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	MySQLHost string
	MySQLPort string
	MySQLUser string
	MySQLPass string
	MySQLDB   string
	AppPort   string
	AppLog    string
}

var C Config

func Load() {
	C = Config{
		MySQLHost: getEnv("MYSQL_HOST", "db"),
		MySQLPort: getEnv("MYSQL_PORT", "3306"),
		MySQLUser: getEnv("MYSQL_USER", "root"),
		MySQLPass: getEnv("MYSQL_PASSWORD", "password"),
		MySQLDB:   getEnv("MYSQL_DATABASE", "ai_forge"),
		AppPort:   getEnv("PORT", "8080"),
		AppLog:    getEnv("APP_LOG_FILE", "./logs/app.log"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
