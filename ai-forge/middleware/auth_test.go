package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
	"github.com/sunyanf/ai-forge/service"
	"gorm.io/gorm"
)

func setupTestDBForMiddleware(t *testing.T) {
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

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDBForMiddleware(t)

	// create user
	u := &model.User{Email: "mw@example.com", PasswordHash: "x", Name: "MW"}
	if err := db.DB.Create(u).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	// generate token
	token, err := service.GenerateJWT(u)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// protected route
	r := gin.New()
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		v, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": v})
	})

	// request with token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body:%s", w.Code, w.Body.String())
	}

	// request without token
	req2 := httptest.NewRequest("GET", "/protected", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", w2.Code)
	}
}
