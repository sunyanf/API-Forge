// Package handler 提供 HTTP 请求处理器的定义。
// 本文件实现健康检查（Health）处理逻辑，
// 用于监控服务运行状态和数据库连接情况，是运维监控的关键接口。
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/response"
)

// HealthResponse 健康检查响应结构体
// 包含服务运行状态、版本信息和数据库连接情况
type HealthResponse struct {
	// Status 服务整体状态：ok 表示正常，degraded 表示降级（部分组件异常）
	Status string `json:"status"`
	// Version 当前服务的版本号
	Version string `json:"version"`
	// Timestamp 健康检查响应的时间戳，格式为 RFC3339
	Timestamp string `json:"timestamp"`
	// Database 数据库连接状态：connected 表示已连接，disconnected 表示断开
	Database string `json:"database"`
}

// Health 返回服务的健康状态
// 会主动探测数据库的连通性，根据探测结果返回 ok 或 degraded 状态。
// 该接口常用于 Kubernetes 的 liveness/readiness probe 或负载均衡器的健康检查。
//
// 请求示例：
//
//	GET /api/health
//
// 响应（数据库正常，200 OK）：
//
//	{"status": "ok", "version": "1.0.0", "timestamp": "...", "database": "connected"}
//
// 响应（数据库断开，503 Service Unavailable）：
//
//	{"status": "degraded", "version": "1.0.0", "timestamp": "...", "database": "disconnected"}
func Health(c *gin.Context) {
	// 步骤1：假设数据库初始状态为断开，后续通过 Ping 操作确认
	dbStatus := "disconnected"

	// 步骤2：检查数据库连通性
	// 通过多层校验确保数据库对象可用
	if db.DB != nil { // 确保数据库对象已初始化
		// 获取底层的 *sql.DB 对象，以执行 Ping 操作
		sqlDB, err := db.DB.DB()
		if err == nil { // 获取底层连接成功
			// Ping 向数据库发送一个轻量级的探测请求，验证连接是否存活
			if err := sqlDB.Ping(); err == nil {
				dbStatus = "connected" // 数据库连通正常
			}
		}
	}

	// 步骤3：构造健康检查响应
	resp := HealthResponse{
		Status:    "ok",                                // 默认状态正常
		Version:   "1.0.0",                             // 服务版本号
		Timestamp: time.Now().UTC().Format(time.RFC3339), // 当前 UTC 时间
		Database:  dbStatus,                            // 数据库连接状态
	}

	// 步骤4：根据数据库状态决定 HTTP 状态码
	// 数据库不可用时返回 503，通知负载均衡器将该实例暂时摘除
	statusCode := http.StatusOK // 默认 200 OK
	if dbStatus != "connected" {
		resp.Status = "degraded"                         // 修改服务状态为降级
		statusCode = http.StatusServiceUnavailable       // 503 Service Unavailable
	}

	// 步骤5：返回 JSON 响应，使用自定义状态码
	response.JSON(c, statusCode, resp)
}
