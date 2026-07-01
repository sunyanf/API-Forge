// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现用户注册（Register）相关的请求处理逻辑。
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

// registerRequest 用户注册请求体
// 用于接收客户端提交的注册表单数据，字段通过 JSON 标签与请求体绑定
type registerRequest struct {
	// Email 用户邮箱地址，必填，需符合邮箱格式
	Email string `json:"email" binding:"required,email"`
	// Password 用户密码，必填，最短长度为 6 个字符
	Password string `json:"password" binding:"required,min=6"`
	// Name 用户昵称/显示名称，选填
	Name string `json:"name"`
}

// Register 处理用户注册请求
// 接收 JSON 格式的注册信息，完成参数校验后调用 service 层创建新用户，
// 成功则返回创建的用户信息（不包含密码），失败则返回错误信息。
//
// 请求示例：
//
//	POST /api/register
//	{"email": "user@example.com", "password": "123456", "name": "Tom"}
//
// 响应（成功，201 Created）：
//
//	{"id": 1, "email": "user@example.com", "name": "Tom", "created_at": "..."}
func Register(c *gin.Context) {
	// 步骤1：将请求体中的 JSON 数据绑定到 registerRequest 结构体
	// ShouldBindJSON 会自动根据 binding 标签进行校验
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 校验失败，返回 400 错误及校验失败的详细原因
		response.BadRequest(c, err.Error())
		return
	}

	// 步骤2：调用 service 层进行用户注册
	// service.RegisterUser 负责密码加密、数据持久化等业务逻辑
	user, err := service.RegisterUser(req.Email, req.Password, req.Name)
	if err != nil {
		// 注册失败（如邮箱已存在），返回错误信息
		response.BadRequest(c, err.Error())
		return
	}

	// 步骤3：构造响应体，将数据库模型映射为对外暴露的响应结构
	// 注意：这里不会返回密码等敏感字段
	resp := UserResponse{
		ID:        user.ID,        // 新用户的自增主键 ID
		Email:     user.Email,     // 用户邮箱
		Name:      user.Name,      // 用户昵称
		CreatedAt: user.CreatedAt, // 注册时间
	}

	// 步骤4：返回 201 Created 状态码及用户信息
	response.Created(c, resp)
}
