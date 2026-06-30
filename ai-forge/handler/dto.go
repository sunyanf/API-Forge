package handler

import "time"

// UserResponse is a safe DTO for /me responses
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	ApiKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
