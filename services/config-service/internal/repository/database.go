package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/agba-ai/config-service/internal/config"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Database represents the database connection
type Database struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*Database, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("driver", cfg.Driver),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.ConnMaxLifetime),
	)

	return &Database{
		db:     db,
		logger: logger,
	}, nil
}

// DB returns the underlying database connection
func (d *Database) DB() *sql.DB {
	return d.db
}

// Ping tests the database connection
func (d *Database) Ping() error {
	return d.db.Ping()
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// BeginTx starts a new transaction
func (d *Database) BeginTx() (*sql.Tx, error) {
	return d.db.Begin()
}

// Health returns the health status of the database
func (d *Database) Health() map[string]interface{} {
	stats := d.db.Stats()
	
	health := map[string]interface{}{
		"status":             "healthy",
		"open_connections":   stats.OpenConnections,
		"in_use":            stats.InUse,
		"idle":              stats.Idle,
		"wait_count":        stats.WaitCount,
		"wait_duration":     stats.WaitDuration.String(),
		"max_idle_closed":   stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	// Check if we can ping the database
	if err := d.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
	}

	return health
}