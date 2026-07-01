// 文件名: admin.go
// 作用: 管理员功能处理器
// 说明: 提供用户管理、角色升级等管理员专属功能。
//       只有角色为 "admin" 的用户可以访问这些接口。

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/response"
	"github.com/sunyanf/ai-forge/service"
)

// AdminUserListResponse 管理员查看的用户列表项
type AdminUserListResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// GetUserList 获取所有用户列表（仅管理员）
// 请求：
//
//	GET /api/v1/admin/users
//	Authorization: Bearer <admin_token>
//
// 响应：用户列表数组
func GetUserList(c *gin.Context) {
	users, err := dao.GetAllUsers()
	if err != nil {
		response.InternalError(c, "获取用户列表失败")
		return
	}

	result := make([]AdminUserListResponse, 0, len(users))
	for _, u := range users {
		result = append(result, AdminUserListResponse{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	response.OK(c, result)
}

// UpgradeUserRequest 升级用户请求
type UpgradeUserRequest struct {
	UserID uint   `json:"user_id" binding:"required"` // 目标用户 ID
	Role   string `json:"role" binding:"required"`    // 新角色：vip / admin / user
}

// UpgradeUser 升级/修改用户角色（仅管理员）
func UpgradeUser(c *gin.Context) {
	var req UpgradeUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 user_id 和 role")
		return
	}

	// 只允许设置有效的角色
	validRoles := map[string]bool{"user": true, "vip": true, "admin": true, "suspended": true}
	if !validRoles[req.Role] {
		response.BadRequest(c, "无效的角色，可选: user, vip, admin, suspended")
		return
	}

	if err := dao.UpdateUserRole(req.UserID, req.Role); err != nil {
		response.InternalError(c, "更新用户角色失败")
		return
	}

	response.OK(c, gin.H{"message": "角色更新成功", "user_id": req.UserID, "new_role": req.Role})
}

// ChangePassword 修改当前用户密码
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6"` // 新密码（≥6位）
}

// ChangePassword 修改当前登录用户的密码
func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 old_password 和 new_password（≥6位）")
		return
	}

	// 从上下文获取当前用户 ID
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	// 验证旧密码
	user, err := service.GetUserByID(userID)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// AuthenticateUser 内部会验证密码哈希
	if _, err := service.AuthenticateUser(user.Email, req.OldPassword); err != nil {
		response.Unauthorized(c, "旧密码不正确")
		return
	}

	// 更新密码
	if err := service.ChangePassword(userID, req.NewPassword); err != nil {
		response.InternalError(c, "密码修改失败")
		return
	}

	response.OK(c, gin.H{"message": "密码修改成功"})
}
