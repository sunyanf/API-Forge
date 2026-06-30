package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

// Register handles user registration
func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := service.RegisterUser(req.Email, req.Password, req.Name)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}
	response.Created(c, resp)
}
