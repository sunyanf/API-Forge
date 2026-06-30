package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/response"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name           string
		dbConnected    bool
		expectedStatus int
		expectedResp   HealthResponse
	}{{
			name:           "database connected",
			dbConnected:    true,
			expectedStatus: http.StatusOK,
			expectedResp: HealthResponse{
				Status:    "ok",
				Version:   "1.0.0",
				Database:  "connected",
			},
		}, {
			name:           "database disconnected",
			dbConnected:    false,
			expectedStatus: http.StatusServiceUnavailable,
			expectedResp: HealthResponse{
				Status:    "degraded",
				Version:   "1.0.0",
				Database:  "disconnected",
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test context
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request
			req, _ := http.NewRequestWithContext(context.Background(), "GET", "/health", nil)
			c.Request = req

			// Mock database connection state
			if tt.dbConnected {
				db.DB = &mockDB{pingError: nil}
			} else {
				db.DB = nil
			}

			// Call handler
			Health(c)

			// Verify response status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var respBody response.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &respBody)
			assert.NoError(t, err)

			// Type assert data to HealthResponse
			healthResp, ok := respBody.Data.(HealthResponse)
			assert.True(t, ok, "response data should be HealthResponse")

			// Verify response fields (excluding timestamp)
			assert.Equal(t, tt.expectedResp.Status, healthResp.Status)
			assert.Equal(t, tt.expectedResp.Version, healthResp.Version)
			assert.Equal(t, tt.expectedResp.Database, healthResp.Database)

			// Verify timestamp is present and valid
			_, err = time.Parse(time.RFC3339, healthResp.Timestamp)
			assert.NoError(t, err, "timestamp should be RFC3339 format")
		})
	}
}

// mockDB is a minimal implementation of *gorm.DB for testing
type mockDB struct {
	pingError error
}

func (m *mockDB) DB() (*db.SqlDB, error) {
	return &mockSqlDB{pingError: m.pingError}, nil
}

// mockSqlDB is a minimal implementation of sql.DB for testing
type mockSqlDB struct {
	pingError error
}

func (m *mockSqlDB) Ping() error {
	return m.pingError
}
