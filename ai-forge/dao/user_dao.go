// 文件名: user_dao.go
// 作用: 用户相关的数据访问层（DAO - Data Access Object）
// 说明: 本层负责与数据库直接交互，封装所有用户表的增删改查操作。
//       上层 service 层通过调用这里的函数来操作数据，不直接接触数据库。

package dao

import (
	"github.com/sunyanf/ai-forge/internal/db" // 数据库连接
	"github.com/sunyanf/ai-forge/model"      // 用户数据模型
)

// CreateUser 在数据库中创建一条新的用户记录
// 参数 u: 待创建的用户对象（指针，创建后 ID 会被自动回填）
// 返回: 可能的数据库错误
func CreateUser(u *model.User) error {
	return db.DB.Create(u).Error
}

// GetUserByEmail 根据邮箱地址查询用户
// 参数 email: 要查询的邮箱地址
// 返回: 用户对象指针和可能的错误
//       如果用户不存在，返回 gorm.ErrRecordNotFound 错误
func GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	// WHERE email = ? 是 GORM 的参数化查询，防止 SQL 注入
	err := db.DB.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
