// 文件名: logging.go
// 作用: 请求日志中间件
// 说明: 记录每个 HTTP 请求的方法、路径、状态码和耗时
//       日志格式示例: GET /api/v1/me 200 3.2ms

package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 请求日志中间件
// 在每个请求结束时，输出一行日志记录
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()         // 记录请求开始时间
		c.Next()                    // 执行后续处理器
		dur := time.Since(start)    // 计算总耗时
		status := c.Writer.Status() // 获取 HTTP 状态码
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, status, dur)
	}
}
