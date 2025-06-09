package repository

import (
    "context"
    "database/sql"
    "go.uber.org/zap"
)

// Repository handles data operations
type Repository struct {
    db     *sql.DB
    logger *zap.Logger
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
    return &Repository{
        db:     db,
        logger: logger,
    }
}

// HealthCheck checks repository health
func (r *Repository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
