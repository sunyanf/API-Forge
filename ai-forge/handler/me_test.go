package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/middleware"
	"github.com/sunyanf/ai-forge/model"
	"github.com/sunyanf/ai-forge/service"
	"gorm.io/gorm"
)

func setupTestDBForMe(t *testing.T) {
	t.Helper()
	conn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.DB = conn
	if err := db.DB.AutoMigrate(&model.User{}, &model.APILog{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
}

func TestMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDBForMe(t)

	// create user
	u := &model.User{Email: "me@example.com", PasswordHash: "x", Name: "Me"}
	if err := db.DB.Create(u).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	// generate token
	token, err := service.GenerateJWT(u)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// router
	r := gin.New()
	r.GET("/api/v1/me", func(c *gin.Context) { c.Next() })
	// attach middleware and handler
	r.GET("/me", middleware.AuthMiddleware(), Me)

	// request with token
	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body:%s", w.Code, w.Body.String())
	}
	// decode response
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["email"] != "me@example.com" {
		t.Fatalf("unexpected email: %v", resp["email"])
	}
	if _, ok := resp["password_hash"]; ok {
		t.Fatalf("password_hash leaked in response")
	}
	if _, ok := resp["PasswordHash"]; ok {
		t.Fatalf("PasswordHash leaked in response")
	}
	if _, ok := resp["id"]; !ok {
		t.Fatalf("id missing")
	}
	if _, ok := resp["created_at"]; !ok {
		t.Fatalf("created_at missing")
	}
}
