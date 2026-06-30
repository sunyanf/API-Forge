package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login and returns JWT token on success
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := service.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, "invalid email or password")
		return
	}
	token, err := service.GenerateJWT(user)
	if err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}
	response.OK(c, gin.H{"token": token})
}
