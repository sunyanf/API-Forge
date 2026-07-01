// 文件名: response.go
// 作用: 统一 API 响应格式
// 说明: 所有 HTTP 响应都经过这里的函数来构造，
//       确保整个项目的返回格式一致：
//       成功 → {"data": ...}
//       失败 → {"error": "...", "message": "..."}
//       这样前端只需要按照统一格式解析即可

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
// 当 API 处理出错时，返回这个结构
type ErrorResponse struct {
	Error   string `json:"error"`             // HTTP 错误类型（如 "Bad Request"）
	Code    string `json:"code,omitempty"`    // 应用层错误码（可选，如 "RATE_LIMIT"）
	Message string `json:"message,omitempty"` // 具体错误描述（给开发者/用户看的）
}

// SuccessResponse 统一成功响应结构
// 所有业务数据都包裹在 data 字段中
type SuccessResponse struct {
	Data    interface{} `json:"data"`              // 响应数据，可以是对象、数组、字符串等
	Message string      `json:"message,omitempty"` // 可选的成功提示消息
}

// JSON 发送任意 HTTP 状态码的成功响应
func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, SuccessResponse{Data: data})
}

// Error 发送统一错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: http.StatusText(code), Message: message})
}

// ErrorWithCode 发送带应用层错误码的统一错误响应
func ErrorWithCode(c *gin.Context, httpCode int, appCode string, message string) {
	c.JSON(httpCode, ErrorResponse{Error: appCode, Message: message})
}

// Created 发送 201 Created 响应（通常用于创建资源成功后）
func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, data)
}

// OK 发送 200 OK 响应（最常用的成功响应）
func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, data)
}

// BadRequest 发送 400 Bad Request 响应（客户端参数错误）
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "请求参数无效"
	}
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 发送 401 Unauthorized 响应（未登录或 Token 无效）
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权"
	}
	Error(c, http.StatusUnauthorized, message)
}

// NotFound 发送 404 Not Found 响应（资源不存在）
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	Error(c, http.StatusNotFound, message)
}

// InternalError 发送 500 Internal Server Error 响应（服务端错误）
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	Error(c, http.StatusInternalServerError, message)
}
