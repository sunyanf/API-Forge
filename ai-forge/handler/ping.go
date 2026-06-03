package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping returns pong
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
