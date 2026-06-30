package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/response"
)

// APIKeyMiddleware authenticates requests using X-API-Key header
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			response.Unauthorized(c, "missing X-API-Key header")
			c.Abort()
			return
		}

		// Look up user by API key
		user, err := dao.GetUserByAPIKey(apiKey)
		if err != nil || user == nil || user.ID == 0 {
			response.Unauthorized(c, "invalid API key")
			c.Abort()
			return
		}

		// Check if user is active (has a valid role)
		if user.Role == "suspended" || user.Role == "disabled" {
			response.Error(c, 403, "account is suspended or disabled")
			c.Abort()
			return
		}

		// Set user context for downstream handlers
		c.Set("user_id", user.ID)
		c.Set("api_key", apiKey)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// APIKeyOrJWTMiddleware tries API Key first, then falls back to JWT Bearer token
func APIKeyOrJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API Key first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := dao.GetUserByAPIKey(apiKey)
			if err == nil && user != nil && user.ID != 0 {
				if user.Role == "suspended" || user.Role == "disabled" {
					response.Error(c, 403, "account is suspended or disabled")
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

		// Fall back to JWT Bearer token
		auth := c.GetHeader("Authorization")
		if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
			// Let the JWT middleware handle this
			// We'll call the JWT middleware manually
			jwtMiddleware := AuthMiddleware()
			jwtMiddleware(c)
			return
		}

		// No valid authentication provided
		response.Unauthorized(c, "authentication required: provide X-API-Key header or Bearer token")
		c.Abort()
	}
}
