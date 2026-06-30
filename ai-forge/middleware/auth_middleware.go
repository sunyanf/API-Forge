package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunyanf/ai-forge/config"
	"github.com/sunyanf/ai-forge/response"
)

// AuthMiddleware verifies Authorization: Bearer <token> and sets user_id in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "missing authorization header",
			})
			return
		}
		var tokenStr string
		if len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		} else {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "invalid authorization header format",
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
				Message: "invalid or expired token",
			})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "invalid token claims",
			})
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "invalid subject",
			})
			return
		}
		uid, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(401, response.ErrorResponse{
				Error:   "Unauthorized",
				Message: "invalid subject",
			})
			return
		}
		c.Set("user_id", uint(uid))
		c.Next()
	}
}
