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

	r := gin.New()
	r.GET("/api/v1/me", middleware.AuthMiddleware(), Me)

	tests := []struct {
		name         string
		setupUser    bool
		expectStatus int
	}{
		{
			name:         "authenticated request returns user info",
			setupUser:    true,
			expectStatus: http.StatusOK,
		},
		{
			name:         "missing user_id from context returns 401",
			setupUser:    false,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "user not found returns 404",
			setupUser:    false,
			expectStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDBForMe(t)

			var token string
			if tt.setupUser {
				u := &model.User{Email: "me@example.com", PasswordHash: "x", Name: "Me"}
				if err := db.DB.Create(u).Error; err != nil {
					t.Fatalf("failed to create user: %v", err)
				}
				var err error
				token, err = service.GenerateJWT(u)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
			} else if tt.expectStatus == http.StatusNotFound {
				// Simulate a valid JWT for a non-existent user
				tok, err := service.GenerateJWT(&model.User{ID: 99999})
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				token = tok
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
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

			if tt.expectStatus == http.StatusOK {
				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("expected 'data' in successful /me response")
				}
				if data["email"] != "me@example.com" {
					t.Errorf("expected email=%q, got %v", "me@example.com", data["email"])
				}
				if _, hasPassword := data["password_hash"]; hasPassword {
					t.Error("password_hash should not be in /me response")
				}
			}
		})
	}
}
