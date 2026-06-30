package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

// Me returns current user info (protected by AuthMiddleware)
func Me(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uid, ok := v.(uint)
	if !ok {
		response.Unauthorized(c, "invalid user id")
		return
	}
	u, err := service.GetUserByID(uid)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	apiKey := ""
	if u.ApiKey != nil {
		apiKey = *u.ApiKey
	}
	resp := UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		ApiKey:    apiKey,
		CreatedAt: u.CreatedAt,
	}
	response.OK(c, resp)
}
