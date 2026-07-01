package config

import (
	"os"
)

// Config 应用配置（从环境变量加载）
type Config struct {
	MySQLHost string // MySQL 主机地址
	MySQLPort string // MySQL 端口
	MySQLUser string // MySQL 用户名
	MySQLPass string // MySQL 密码
	MySQLDB   string // 数据库名称
	AppPort   string // 服务端口
	AppLog    string // 日志文件路径
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
