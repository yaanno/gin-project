package sqlite

import (
	"log"
	"time"

	"github.com/yourusername/user-management-api/internal/database/migrations"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Path            string
	InMemory        bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func InitializeDatabase(config DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, apperrors.NewInitializationError("Failed to initialize database", err)
	}

	// Connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, apperrors.NewInitializationError("Failed to get database connection", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)       // Maximum number of open connections
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)       // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime) // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime) // Maximum idle time before closing

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
		return nil, apperrors.NewInitializationError("Database migration failed", err)
	}

	return db, nil
}
