// Package handler 提供 HTTP 请求处理器的定义。
// 本文件定义处理层的数据传输对象（DTO，Data Transfer Object），
// 用于在各处理器间共享通用的请求/响应结构体。
package handler

import "time"

// UserResponse 用户信息响应体
// 该结构体用于所有返回用户信息的接口（注册、登录、获取个人信息等）。
// 只包含对外公开的字段，不会泄露密码哈希等敏感数据。
type UserResponse struct {
	// ID 用户唯一标识，数据库自增主键
	ID uint `json:"id"`
	// Email 用户邮箱地址
	Email string `json:"email"`
	// Name 用户昵称/显示名称
	Name string `json:"name"`
	// Role 用户角色，可取值为 user（普通用户）、vip（VIP 用户）、admin（管理员）
	Role string `json:"role"`
	// ApiKey 用户的 API 访问密钥，omitempty 表示该字段为空时不返回
	ApiKey string `json:"api_key,omitempty"`
	// CreatedAt 账户注册时间
	CreatedAt time.Time `json:"created_at"`
}
