package database

import (
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection creates a new database connection
func NewConnection(dsn string, log *slog.Logger) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("database connection established")
	return db, nil
}

// AutoMigrate runs automatic migration for all entities
func AutoMigrate(db *gorm.DB, log *slog.Logger) error {
	log.Info("running database migrations")

	if err := db.AutoMigrate(
		&entities.User{},
		&entities.Product{},
		&entities.Transaction{},
		&entities.TransactionItem{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Info("database migrations completed")
	return nil
}
