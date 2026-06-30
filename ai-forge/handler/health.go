package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/response"
)

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

// Health returns detailed health status including database connectivity
func Health(c *gin.Context) {
	dbStatus := "disconnected"

	// Check database connectivity
	if db.DB != nil {
		sqlDB, err := db.DB.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				dbStatus = "connected"
			}
		}
	}

	resp := HealthResponse{
		Status:    "ok",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  dbStatus,
	}

	statusCode := http.StatusOK
	if dbStatus != "connected" {
		resp.Status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	response.JSON(c, statusCode, resp)
}
