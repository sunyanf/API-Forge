// 文件名: api_log_dao.go
// 作用: API 调用日志的数据访问层（DAO）

package dao

import (
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
)

// CreateLog 创建一条 API 调用日志记录
// 参数 l: 待创建的日志对象（指针）
// 返回: 可能的数据库错误
func CreateLog(l *model.APILog) error {
	return db.DB.Create(l).Error
}
