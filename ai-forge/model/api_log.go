// 文件名: api_log.go
// 作用: 定义 API 调用日志的数据模型
// 说明: 记录每一次 API 调用的详细信息，用于统计、审计和问题排查

package model

import "time"

// APILog API 调用日志记录
// 每次用户通过 API Key 调用服务时，系统会自动创建一条日志记录
type APILog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`                   // 日志ID（主键）
	ProjectID       uint      `json:"project_id"`                            // 所属项目ID
	Endpoint        string    `gorm:"size:255" json:"endpoint"`              // 调用的 API 路径
	ApiKey          string    `gorm:"size:64" json:"api_key"`                // 使用的 API Key
	ClientIP        string    `gorm:"size:64" json:"client_ip"`              // 客户端 IP 地址
	RequestPayload  string    `gorm:"type:longtext" json:"request_payload"`   // 请求体内容
	ResponsePayload string    `gorm:"type:longtext" json:"response_payload"`  // 响应体内容
	Status          string    `gorm:"size:32" json:"status"`                 // 响应状态：success / error
	TokensIn        int       `json:"tokens_in"`                             // 输入 Token 数量
	TokensOut       int       `json:"tokens_out"`                            // 输出 Token 数量
	DurationMs      int       `json:"duration_ms"`                           // 请求耗时（毫秒）
	CreatedAt       time.Time `json:"created_at"`                            // 请求时间
}
