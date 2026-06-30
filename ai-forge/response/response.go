package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the unified error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse is the unified success response structure
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// JSON sends a unified success response
func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, SuccessResponse{Data: data})
}

// Error sends a unified error response
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: http.StatusText(code), Message: message})
}

// ErrorWithCode sends a unified error response with an application-level error code
func ErrorWithCode(c *gin.Context, httpCode int, appCode string, message string) {
	c.JSON(httpCode, ErrorResponse{Error: appCode, Message: message})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, data)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "invalid request parameters"
	}
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	Error(c, http.StatusUnauthorized, message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "resource not found"
	}
	Error(c, http.StatusNotFound, message)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	Error(c, http.StatusInternalServerError, message)
}
