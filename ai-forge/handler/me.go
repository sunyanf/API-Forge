// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现获取当前登录用户信息（Me）的请求处理逻辑。
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

// Me 获取当前登录用户的信息（需 JWT 鉴权）
// 该接口依赖 JWT 中间件将解析后的 user_id 注入到请求上下文中。
// 根据上下文中的用户 ID 查询数据库，返回用户的公开信息（不含密码）。
//
// 调用前提：客户端必须在 Authorization 头中携带有效的 JWT Token，
// JWT 中间件会在到达本处理器之前完成 Token 解析并将 user_id 写入上下文。
//
// 请求示例：
//
//	GET /api/me
//	Authorization: Bearer <JWT Token>
//
// 响应（成功，200 OK）：
//
//	{"id": 1, "email": "user@example.com", "name": "Tom", "role": "user", ...}
//
// 响应（失败，401 Unauthorized）：
//
//	{"message": "未登录"}
func Me(c *gin.Context) {
	// 步骤1：从请求上下文中获取由 JWT 中间件注入的 user_id
	// 使用类型断言检查值是否存在
	v, ok := c.Get("user_id")
	if !ok {
		// 上下文中没有 user_id，说明请求未经过 JWT 中间件或 Token 无效
		response.Unauthorized(c, "未登录")
		return
	}

	// 步骤2：将 interface{} 类型断言为 uint（数据库主键类型）
	uid, ok := v.(uint)
	if !ok {
		// 类型不匹配，说明中间件写入的数据格式异常
		response.Unauthorized(c, "用户ID无效")
		return
	}

	// 步骤3：根据用户 ID 从数据库查询用户信息
	u, err := service.GetUserByID(uid)
	if err != nil {
		// 用户不存在（可能已被删除），返回 404
		response.NotFound(c, "用户不存在")
		return
	}

	// 步骤4：安全地处理指针类型的 ApiKey 字段
	// ApiKey 可能为 nil（用户尚未生成 API Key），需要做空值判断
	apiKey := ""
	if u.ApiKey != nil {
		apiKey = *u.ApiKey // 解引用指针，获取实际的 API Key 字符串
	}

	// 步骤5：构造响应体，将数据库模型映射为对外暴露的结构
	// 注意：不会返回密码哈希等敏感字段
	resp := UserResponse{
		ID:        u.ID,        // 用户唯一标识
		Email:     u.Email,     // 用户邮箱
		Name:      u.Name,      // 用户昵称
		Role:      u.Role,      // 用户角色（user/vip/admin）
		ApiKey:    apiKey,      // API 密钥（可能为空）
		CreatedAt: u.CreatedAt, // 注册时间
	}

	// 步骤6：返回 200 OK 及用户信息
	response.OK(c, resp)
}
