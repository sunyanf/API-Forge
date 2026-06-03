package dao

import (
	"github.com/sunyanf/ai-forge/internal/db"
	"github.com/sunyanf/ai-forge/model"
)

func CreateLog(l *model.APILog) error {
	return db.DB.Create(l).Error
}
