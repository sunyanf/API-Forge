// 文件名: apikey_middleware.go
// 作用: API Key 鉴权中间件
// 说明: 支持两种认证方式——X-API-Key 请求头和 JWT Bearer Token。
//       外部 API 调用者使用 X-API-Key，前端用户使用 JWT Token。

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/response"
)

// APIKeyMiddleware 使用 X-API-Key 请求头鉴权
// 适用于外部程序通过 API Key 调用服务的场景
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			response.Unauthorized(c, "缺少 X-API-Key 请求头")
			c.Abort()
			return
		}

		// 根据 API Key 查找用户
		user, err := dao.GetUserByAPIKey(apiKey)
		if err != nil || user == nil || user.ID == 0 {
			response.Unauthorized(c, "API Key 无效")
			c.Abort()
			return
		}

		// 检查账户状态
		if user.Role == "suspended" || user.Role == "disabled" {
			response.Error(c, 403, "账户已被暂停或禁用")
			c.Abort()
			return
		}

		// 将用户信息注入上下文
		c.Set("user_id", user.ID)
		c.Set("api_key", apiKey)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// APIKeyOrJWTMiddleware 先尝试 API Key，然后回退到 JWT
// 适用于既支持外部 API 调用也支持前端登录的接口
func APIKeyOrJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 第一步：尝试 X-API-Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := dao.GetUserByAPIKey(apiKey)
			if err == nil && user != nil && user.ID != 0 {
				if user.Role == "suspended" || user.Role == "disabled" {
					response.Error(c, 403, "账户已被暂停或禁用")
					c.Abort()
					return
				}
				c.Set("user_id", user.ID)
				c.Set("api_key", apiKey)
				c.Set("user_role", user.Role)
				c.Next()
				return
			}
		}

		// 第二步：回退到 JWT Bearer Token
		auth := c.GetHeader("Authorization")
		if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
			jwtMiddleware := AuthMiddleware()
			jwtMiddleware(c)
			return
		}

		// 第三步：都没有则拒绝
		response.Unauthorized(c, "需要认证：请提供 X-API-Key 或 Bearer Token")
		c.Abort()
	}
}
