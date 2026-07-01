// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现连通性测试（Ping）处理器，
// 是最简单的接口，用于快速验证服务是否在线和可达。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping 处理连通性测试请求，返回 "pong"
// 这是一个最简单的健康探测接口，仅验证 HTTP 服务是否正常响应。
// 与 /api/health 不同，此接口不检查数据库等外部依赖，只确认服务进程在运行。
//
// 用途：
//   - 快速验证服务是否启动
//   - 负载均衡器的基础存活检测
//   - 前端调用前的连通性预检
//
// 请求示例：
//
//	GET /api/ping
//
// 响应（200 OK）：
//
//	{"message": "pong"}
func Ping(c *gin.Context) {
	// 直接返回 200 OK 和 {"message": "pong"} JSON 响应
	// gin.H 是 GIN 框架提供的快捷方式，等同于 map[string]interface{}
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
