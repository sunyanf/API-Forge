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

	r := gin.New()
	r.POST("/api/v1/login", Login)

	tests := []struct {
		name         string
		email        string
		password     string
		expectStatus int
		setupUser    bool // create a user before this test
	}{
		{
			name:         "valid login returns JWT token",
			email:        "login@example.com",
			password:     "secret123",
			expectStatus: http.StatusOK,
			setupUser:    true,
		},
		{
			name:         "wrong password returns 401",
			email:        "login@example.com",
			password:     "wrongpassword",
			expectStatus: http.StatusUnauthorized,
			setupUser:    true,
		},
		{
			name:         "non-existent user returns 401",
			email:        "ghost@example.com",
			password:     "secret123",
			expectStatus: http.StatusUnauthorized,
			setupUser:    false,
		},
		{
			name:         "missing email field returns 400",
			email:        "",
			password:     "secret123",
			expectStatus: http.StatusBadRequest,
			setupUser:    false,
		},
		{
			name:         "invalid email format returns 400",
			email:        "not-an-email",
			password:     "secret123",
			expectStatus: http.StatusBadRequest,
			setupUser:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDBForLogin(t)

			// Pre-insert user if needed
			if tt.setupUser && tt.email != "" {
				pw, _ := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.DefaultCost)
				u := &model.User{Email: tt.email, PasswordHash: string(pw), Name: "Login"}
				if err := db.DB.Create(u).Error; err != nil {
					t.Fatalf("failed to create user: %v", err)
				}
			}

			payload := map[string]string{"email": tt.email, "password": tt.password}
			b, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("status = %d; want %d, body: %s", w.Code, tt.expectStatus, w.Body.String())
			}

			// Verify response envelope
			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}

			// For successful login, verify token in data
			if tt.expectStatus == http.StatusOK {
				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("expected 'data' in successful login response")
				}
				token, hasToken := data["token"]
				if !hasToken || token == "" {
					t.Fatal("expected non-empty 'token' in successful login response")
				}
			}
		})
	}
}
