package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/service"
)

// UserProfile returns current user info including api key (protected)
func UserProfile(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
		return
	}
	uid, ok := v.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}
	u, err := service.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
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
	c.JSON(http.StatusOK, resp)
}
