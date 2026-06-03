package model

import "time"

// APILog records API invocations
type APILog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectID       uint      `json:"project_id"`
	Endpoint        string    `gorm:"size:255" json:"endpoint"`
	ApiKey          string    `gorm:"size:64" json:"api_key"`
	ClientIP        string    `gorm:"size:64" json:"client_ip"`
	RequestPayload  string    `gorm:"type:longtext" json:"request_payload"`
	ResponsePayload string    `gorm:"type:longtext" json:"response_payload"`
	Status          string    `gorm:"size:32" json:"status"`
	TokensIn        int       `json:"tokens_in"`
	TokensOut       int       `json:"tokens_out"`
	DurationMs      int       `json:"duration_ms"`
	CreatedAt       time.Time `json:"created_at"`
}
