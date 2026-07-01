// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现用户登录（Login）相关的请求处理逻辑。
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

// loginRequest 用户登录请求体
// 用于接收客户端提交的登录凭据
type loginRequest struct {
	// Email 用户邮箱地址，必填，需符合邮箱格式
	Email string `json:"email" binding:"required,email"`
	// Password 用户密码，必填
	Password string `json:"password" binding:"required"`
}

// Login 处理用户登录请求
// 接收邮箱和密码，验证通过后生成 JWT Token 并返回。
// 客户端需在后续请求的 Authorization 头中携带该 Token。
//
// 请求示例：
//
//	POST /api/login
//	{"email": "user@example.com", "password": "123456"}
//
// 响应（成功，200 OK）：
//
//	{"token": "eyJhbGciOiJIUzI1NiIs..."}
//
// 响应（失败，401 Unauthorized）：
//
//	{"message": "用户名或密码错误"}
func Login(c *gin.Context) {
	// 步骤1：绑定并校验请求体中的 JSON 数据
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 校验失败（如邮箱格式不对、字段缺失），返回 400 错误
		response.BadRequest(c, err.Error())
		return
	}

	// 步骤2：调用 service 层验证用户凭据
	// AuthenticateUser 会查询数据库并比对密码哈希
	user, err := service.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		// 邮箱不存在或密码不匹配，返回 401 未授权
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 步骤3：认证成功后，为当前用户生成 JWT Token
	// Token 中通常包含用户 ID、角色等声明（claims），并设有过期时间
	token, err := service.GenerateJWT(user)
	if err != nil {
		// Token 生成异常（如密钥配置问题），返回 500 内部错误
		response.InternalError(c, "Token 生成失败")
		return
	}

	// 步骤4：返回生成的 JWT Token
	// 客户端应将其保存在本地（如 localStorage），后续请求放在 Authorization 头中
	response.OK(c, gin.H{"token": token})
}
