package testhelper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupTestGin sets up a Gin test context
func SetupTestGin(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// AssertJSONResponse asserts that the response is valid JSON with expected status code
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) map[string]interface{} {
	t.Helper()
	assert.Equal(t, expectedCode, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "response should be valid JSON")
	return resp
}

// AssertErrorResponse asserts error response structure
func AssertErrorResponse(t *testing.T, resp map[string]interface{}, expectedError string) {
	t.Helper()
	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "expected data wrapper")
	errorMsg, ok := data["error"].(string)
	assert.True(t, ok, "expected error field")
	assert.Contains(t, errorMsg, expectedError, "error message should contain expected text")
}

// AssertSuccessResponse asserts success response structure
func AssertSuccessResponse(t *testing.T, resp map[string]interface{}, expectedFields []string) {
	t.Helper()
	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "expected data wrapper")
	for _, field := range expectedFields {
		_, hasField := data[field]
		assert.True(t, hasField, "expected field %q in response data", field)
	}
}