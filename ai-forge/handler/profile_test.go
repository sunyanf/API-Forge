package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/middleware"
	"github.com/sunyanf/ai-forge/model"
	"github.com/sunyanf/ai-forge/service"
	"gorm.io/gorm"
)

func setupTestDBForProfile(t *testing.T) {
	t.Helper()
	conn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.DB = conn
	if err := db.DB.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
}

func TestUserProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDBForProfile(t)

	// create user with api key
	apiKey := "test-api-key-123"
	u := &model.User{Email: "profile@example.com", PasswordHash: "x", Name: "Profile", ApiKey: &apiKey}
	if err := db.DB.Create(u).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	// generate token
	token, err := service.GenerateJWT(u)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	r := gin.New()
	r.GET("/user/profile", middleware.AuthMiddleware(), UserProfile)

	req := httptest.NewRequest("GET", "/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body:%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["api_key"] != "test-api-key-123" {
		t.Fatalf("unexpected api_key: %v", resp["api_key"])
	}
}
