package db

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

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

// Connect connects to MySQL using GORM. Prefer MYSQL_DSN if set, otherwise
// build DSN from MYSQL_USER/MYSQL_PASSWORD/MYSQL_HOST/MYSQL_PORT/MYSQL_DATABASE.
// It also configures GORM logger to write migration logs to APP_LOG_FILE.
// Connect will retry until database becomes available (useful for container startup).
func Connect() error {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		user := getenv("MYSQL_USER", "root")
		pass := getenv("MYSQL_PASSWORD", "password")
		host := getenv("MYSQL_HOST", "127.0.0.1")
		port := getenv("MYSQL_PORT", "3306")
		dbname := getenv("MYSQL_DATABASE", "ai_forge")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)
	}

	// prepare log output for GORM
	logPath := getenv("APP_LOG_FILE", "./logs/app.log")
	if err := ensureDir(filepath.Dir(logPath)); err != nil {
		return fmt.Errorf("ensure log dir: %w", err)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	// also write GORM logs to stdout
	writer := io.MultiWriter(os.Stdout, f)
	newLogger := logger.New(log.New(writer, "", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})

	// retry loop
	var dbConn *gorm.DB
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
