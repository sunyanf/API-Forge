package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sunyanf/ai-forge/config"
	rootHandler "github.com/sunyanf/ai-forge/handler"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/internal/router"
	"github.com/sunyanf/ai-forge/middleware"
	"github.com/sunyanf/ai-forge/model"
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

	// Health check endpoint
	r.GET("/health", rootHandler.Health)

	// Public API endpoints
	api := r.Group("/api/v1")
	{
		// Authentication endpoints (public)
		api.POST("/register", rootHandler.Register)
		api.POST("/login", rootHandler.Login)

		// Protected user endpoints
		api.GET("/me", middleware.AuthMiddleware(), rootHandler.Me)

		// Free service endpoints (with rate limiting)
		api.GET("/ip/location", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetIPLocation)
		api.GET("/image/random", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetRandomImage)
		api.GET("/image/redirect", middleware.RateLimitMiddleware(60, time.Minute), rootHandler.GetRandomImageRedirect)

		// API Key protected endpoints (for external API access)
		apiWithKey := api.Group("")
		apiWithKey.Use(middleware.APIKeyOrJWTMiddleware())
		{
			// Add API key protected endpoints here
			// apiWithKey.GET("/service/example", rootHandler.ExampleService)
		}
	}

	// serve OpenAPI spec and simple Swagger UI index
	r.StaticFile("/docs/openapi.yaml", "./openapi.yaml")
	r.GET("/docs", func(c *gin.Context) { c.File("./docs/swagger_index.html") })

	port := config.C.AppPort
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
