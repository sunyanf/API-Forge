package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/internal/handler"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/ping", handler.Ping)
}
