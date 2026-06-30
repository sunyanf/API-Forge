package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRandomImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/v1/random-image", GetRandomImage)

	tests := []struct {
		name         string
		query        string
		expectStatus int
		checkData    func(t *testing.T, body []byte)
	}{
		{
			name:         "valid dimensions with defaults",
			query:        "",
			expectStatus: http.StatusOK,
			checkData: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("missing data envelope")
				}
				// Defaults should be applied
				if data["width"] != float64(1280) {
					t.Errorf("expected default width 1280, got %v", data["width"])
				}
				if data["height"] != float64(720) {
					t.Errorf("expected default height 720, got %v", data["height"])
				}
				if data["provider"] != "picsum.photos" {
					t.Errorf("unexpected provider: %v", data["provider"])
				}
			},
		},
		{
			name:         "custom valid dimensions",
			query:        "width=800&height=600",
			expectStatus: http.StatusOK,
			checkData: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("missing data envelope")
				}
				if data["width"] != float64(800) {
					t.Errorf("expected width 800, got %v", data["width"])
				}
				if data["height"] != float64(600) {
					t.Errorf("expected height 600, got %v", data["height"])
				}
			},
		},
		{
			name:         "valid category accepted",
			query:        "category=nature",
			expectStatus: http.StatusOK,
			checkData: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("missing data envelope")
				}
				if data["category"] != "nature" {
					t.Errorf("expected category 'nature', got %v", data["category"])
				}
			},
		},
		{
			name:         "case-insensitive category",
			query:        "category=NATURE",
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid category rejected",
			query:        "category=nonexistent",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "width below minimum rejected",
			query:        "width=50",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "width above maximum rejected",
			query:        "width=5000",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "height below minimum rejected",
			query:        "height=50",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "height above maximum rejected",
			query:        "height=5000",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "exact boundary width=100 accepted",
			query:        "width=100&height=100",
			expectStatus: http.StatusOK,
		},
		{
			name:         "exact boundary width=4096 accepted",
			query:        "width=4096&height=4096",
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/random-image?"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("status = %d; want %d, body: %s", w.Code, tt.expectStatus, w.Body.String())
			}

			if tt.checkData != nil {
				tt.checkData(t, w.Body.Bytes())
			}
		})
	}
}
