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

	r := gin.New()
	r.POST("/api/v1/register", Register)

	tests := []struct {
		name         string
		payload      map[string]string
		expectStatus int
		setupUser    bool // pre-insert a user with this email to test duplicate
	}{
		{
			name:         "valid registration",
			payload:      map[string]string{"email": "new@example.com", "password": "secret123", "name": "New User"},
			expectStatus: http.StatusCreated,
		},
		{
			name:         "missing email field",
			payload:      map[string]string{"password": "secret123", "name": "No Email"},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "invalid email format",
			payload:      map[string]string{"email": "not-an-email", "password": "secret123", "name": "Bad Email"},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "password shorter than 6 chars",
			payload:      map[string]string{"email": "short@example.com", "password": "abc", "name": "Short PW"},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "duplicate email registration",
			payload:      map[string]string{"email": "dup@example.com", "password": "secret123", "name": "Duplicate"},
			expectStatus: http.StatusBadRequest,
			setupUser:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// fresh DB per sub-test
			setupTestDB(t)

			// Pre-insert existing user if needed
			if tt.setupUser {
				existing := &model.User{Email: tt.payload["email"], Name: "Existing"}
				if err := db.DB.Create(existing).Error; err != nil {
					t.Fatalf("failed to create existing user: %v", err)
				}
			}

			b, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("status = %d; want %d, body: %s", w.Code, tt.expectStatus, w.Body.String())
			}

			// Verify the response is a proper JSON envelope
			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}

			// For success cases, verify data envelope contains expected fields
			if tt.expectStatus == http.StatusCreated {
				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("expected 'data' in response for successful registration")
				}
				if _, hasID := data["id"]; !hasID {
					t.Error("expected 'id' in response data")
				}
				if data["email"] != tt.payload["email"] {
					t.Errorf("expected email=%q, got %v", tt.payload["email"], data["email"])
				}
			}
		})
	}
}
