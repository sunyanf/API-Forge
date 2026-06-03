package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDBForLogin(t *testing.T) {
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

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDBForLogin(t)

	// create user
	pw, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	u := &model.User{Email: "login@example.com", PasswordHash: string(pw), Name: "Login"}
	if err := db.DB.Create(u).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// setup router
	r := gin.New()
	r.POST("/api/v1/login", Login)

	payload := map[string]string{"email": "login@example.com", "password": "secret123"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body:%s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["token"] == "" {
		t.Fatalf("expected token in response, got empty")
	}
}
