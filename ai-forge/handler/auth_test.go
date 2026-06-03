package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
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

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.POST("/api/v1/register", Register)

	payload := map[string]string{"email": "test@example.com", "password": "secret123", "name": "Test"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body:%s", w.Code, w.Body.String())
	}

	// verify user in DB
	var u model.User
	if err := db.DB.Where("email = ?", "test@example.com").First(&u).Error; err != nil {
		t.Fatalf("user not found in db: %v", err)
	}

	// duplicate registration should fail
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code == http.StatusCreated {
		t.Fatalf("expected duplicate to fail, got %d", w2.Code)
	}
}
