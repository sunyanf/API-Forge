// 文件名: auth_middleware.go
// 作用: JWT 鉴权中间件
// 说明: 验证请求头中的 Bearer Token，解析出 user_id 和 user_role 注入上下文

package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunyanf/ai-forge/config"
	"github.com/sunyanf/ai-forge/response"
)

// AuthMiddleware JWT 鉴权中间件，验证 Authorization: Bearer <token> 并将 user_id 和 user_role 存入上下文
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "缺少 Authorization 请求头",
			})
			return
		}
		var tokenStr string
		if len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		} else {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authorization 头格式错误，应为 Bearer <token>",
			})
			return
		}
		secret := config.JWTSecret()
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Token 无效或已过期",
			})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Token 声明无效",
			})
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Token 主题无效",
			})
			return
		}
		uid, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Token 主题无效",
			})
			return
		}
		c.Set("user_id", uint(uid))
		// 注入用户角色，用于后续 VIP 权限校验
		role, _ := claims["role"].(string)
		if role == "" {
			role = "user"
		}
		c.Set("user_role", role)
		c.Next()
	}
}
