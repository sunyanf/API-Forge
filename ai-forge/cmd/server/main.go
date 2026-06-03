package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/sunyanf/ai-forge/config"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
	"github.com/sunyanf/ai-forge/internal/router"
	rootHandler "github.com/sunyanf/ai-forge/handler"
	"github.com/sunyanf/ai-forge/middleware"
)

func ensureLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func main() {
	// load config
	config.Load()

	// configure log file
	logPath := config.C.AppLog
	f, err := ensureLogFile(logPath)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	// also let gin write to the same file and stdout
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(f, os.Stderr)

	// connect to DB
	if err := db.Connect(); err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	// automigrate
	if err := db.AutoMigrate(&model.User{}, &model.APILog{}); err != nil {
		log.Fatalf("db automigrate failed: %v", err)
	}

	r := gin.Default()
	r.Use(middleware.RequestLogger())
	router.RegisterRoutes(r)

	api := r.Group("/api/v1")
	{
		api.POST("/register", rootHandler.Register)
		api.POST("/login", rootHandler.Login)
		api.GET("/me", middleware.AuthMiddleware(), rootHandler.Me)
	}

	port := config.C.AppPort
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
