package repository

import (
	"context"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type auditLogRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewAuditLogRepository creates a new audit log repository.
func NewAuditLogRepository(db *gorm.DB, logger *slog.Logger) interfaces.AuditLogRepository {
	return &auditLogRepository{db: db, logger: logger}
}

// Create inserts a new audit log row.
func (r *auditLogRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create audit log", "error", err)
		return err
	}
	return nil
}
