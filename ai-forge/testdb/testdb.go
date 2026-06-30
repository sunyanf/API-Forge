package testdb

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// AutoMigrateModels migrates the given models to the test database
func AutoMigrateModels(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}