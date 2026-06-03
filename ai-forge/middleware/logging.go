package middleware

import (
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

// Simple request logger middleware that logs to the configured log output
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		status := c.Writer.Status()
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, status, dur)
	}
}
