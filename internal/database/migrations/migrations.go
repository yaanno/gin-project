package migrations

import (
	"log"

	"github.com/yourusername/user-management-api/internal/database"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes
	err := db.AutoMigrate(
		&database.User{},
		&database.LoginAttempt{},
	)

	if err != nil {
		return err
	}

	// Custom index creation (if not automatically handled by GORM)
	if err := createCustomIndexes(db); err != nil {
		return err
	}

	return nil
}

func createCustomIndexes(db *gorm.DB) error {
	// User indexes
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_user_status ON users(status);
		CREATE INDEX IF NOT EXISTS idx_user_last_activity ON users(last_activity_at);
	`).Error; err != nil {
		log.Printf("Error creating user indexes: %v", err)
		return err
	}

	return nil
}
