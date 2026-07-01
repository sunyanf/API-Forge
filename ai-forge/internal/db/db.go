package db

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// Connect 连接数据库。默认使用 SQLite（零依赖本地开发），
// 设置 DB_TYPE=mysql 则使用 MySQL（需运行中的 MySQL 实例）。
// 同时将 GORM 日志输出到 APP_LOG_FILE。
// 连接 MySQL 时会自动重试直到数据库可用（适用于容器启动场景）。
func Connect() error {
	// DB_TYPE 默认为 "sqlite"，实现零依赖本地开发
	// 设置 DB_TYPE=mysql 切换到 MySQL（需要在 Docker 中运行）
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	// 准备 GORM 日志输出
	logPath := getenv("APP_LOG_FILE", "./logs/app.log")
	if err := ensureDir(filepath.Dir(logPath)); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	// 同时写入 stdout
	writer := io.MultiWriter(os.Stdout, f)
	newLogger := logger.New(log.New(writer, "", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})

	var dbConn *gorm.DB

	// SQLite 模式
	if dbType == "sqlite" {
		dbFile := getenv("SQLITE_FILE", "./data.db")
		dbConn, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{Logger: newLogger})
		if err != nil {
			return fmt.Errorf("无法打开 SQLite 数据库: %w", err)
		}
		DB = dbConn
		log.Printf("[db] 已连接 SQLite: %s", dbFile)
		return nil
	}

	// MySQL 模式
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		user := getenv("MYSQL_USER", "root")
		pass := getenv("MYSQL_PASSWORD", "password")
		host := getenv("MYSQL_HOST", "127.0.0.1")
		port := getenv("MYSQL_PORT", "3306")
		dbname := getenv("MYSQL_DATABASE", "ai_forge")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)
	}

	// retry loop
	var lastErr error
	for i := 0; i < 60; i++ {
		dbConn, lastErr = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
		if lastErr == nil {
			// ping sql.DB
			sqlDB, err := dbConn.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					DB = dbConn
					log.Printf("[db] connected after %d attempts", i+1)
					return nil
				}
				lastErr = err
			}
		}
		log.Printf("[db] connect attempt %d failed: %v", i+1, lastErr)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("failed to initialize database, got error %v", lastErr)
}

// AutoMigrate runs GORM automigrations for given models
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("db not connected")
	}
	return DB.AutoMigrate(models...)
}
