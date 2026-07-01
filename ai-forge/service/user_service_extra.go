// 文件名: user_service_extra.go
// 作用: 用户相关业务逻辑的补充函数

package service

import (
	"github.com/sunyanf/ai-forge/dao"
	"github.com/sunyanf/ai-forge/model"
)

// GetUserByID 根据用户 ID 查询用户信息
// 参数 id: 用户主键 ID
// 返回: 用户对象和可能的错误（用户不存在时返回 nil 和 gorm.ErrRecordNotFound）
func GetUserByID(id uint) (*model.User, error) {
	return dao.GetUserByID(id)
}
