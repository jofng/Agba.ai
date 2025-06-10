package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"config-service/internal/config"
	"config-service/internal/models"
	"github.com/google/uuid"
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

// DatabaseRepository implements configuration repository operations
type DatabaseRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewDatabaseRepository creates a new database repository
func NewDatabaseRepository(db *Database, logger *zap.Logger) *DatabaseRepository {
	return &DatabaseRepository{
		db:     db,
		logger: logger,
	}
}

// CreateConfig creates a new configuration in the database
func (r *DatabaseRepository) CreateConfig(ctx context.Context, config *models.Configuration) error {
	query := `
		INSERT INTO configurations (
			id, service, environment, version, data, schema, tags, metadata,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	dataJSON, err := json.Marshal(config.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	schemaJSON, err := json.Marshal(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal config schema: %w", err)
	}

	tagsJSON, err := json.Marshal(config.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal config tags: %w", err)
	}

	metadataJSON, err := json.Marshal(config.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal config metadata: %w", err)
	}

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	_, err = r.db.DB().ExecContext(ctx, query,
		config.ID, config.Service, config.Environment, config.Version,
		dataJSON, schemaJSON, tagsJSON, metadataJSON,
		config.CreatedBy, config.UpdatedBy, config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create configuration", zap.Error(err))
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	r.logger.Info("Configuration created successfully", zap.String("id", config.ID.String()))
	return nil
}

// GetConfig retrieves a configuration by ID
func (r *DatabaseRepository) GetConfig(ctx context.Context, id uuid.UUID) (*models.Configuration, error) {
	query := `
		SELECT id, service, environment, version, data, schema, tags, metadata,
			   created_by, updated_by, created_at, updated_at
		FROM configurations 
		WHERE id = $1`

	var config models.Configuration
	var dataJSON, schemaJSON, tagsJSON, metadataJSON []byte

	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(
		&config.ID, &config.Service, &config.Environment, &config.Version,
		&dataJSON, &schemaJSON, &tagsJSON, &metadataJSON,
		&config.CreatedBy, &config.UpdatedBy, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("configuration not found: %s", id.String())
		}
		r.logger.Error("Failed to get configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(dataJSON, &config.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}
	if err := json.Unmarshal(schemaJSON, &config.Schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config schema: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &config.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config tags: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config metadata: %w", err)
	}

	return &config, nil
}

// GetConfigs retrieves configurations with pagination and filtering
func (r *DatabaseRepository) GetConfigs(ctx context.Context, page, limit int, service, environment string) ([]*models.Configuration, int64, error) {
	offset := (page - 1) * limit
	
	// Build query with optional filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if service != "" {
		whereClause += fmt.Sprintf(" AND service = $%d", argIndex)
		args = append(args, service)
		argIndex++
	}
	if environment != "" {
		whereClause += fmt.Sprintf(" AND environment = $%d", argIndex)
		args = append(args, environment)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM configurations %s", whereClause)
	var total int64
	err := r.db.DB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count configurations: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, service, environment, version, data, schema, tags, metadata,
			   created_by, updated_by, created_at, updated_at
		FROM configurations %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)

	rows, err := r.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query configurations: %w", err)
	}
	defer rows.Close()

	var configs []*models.Configuration
	for rows.Next() {
		var config models.Configuration
		var dataJSON, schemaJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&config.ID, &config.Service, &config.Environment, &config.Version,
			&dataJSON, &schemaJSON, &tagsJSON, &metadataJSON,
			&config.CreatedBy, &config.UpdatedBy, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan configuration: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(dataJSON, &config.Data); err != nil {
			r.logger.Warn("Failed to unmarshal config data", zap.Error(err))
			config.Data = make(map[string]interface{})
		}
		if err := json.Unmarshal(schemaJSON, &config.Schema); err != nil {
			r.logger.Warn("Failed to unmarshal config schema", zap.Error(err))
			config.Schema = make(map[string]interface{})
		}
		if err := json.Unmarshal(tagsJSON, &config.Tags); err != nil {
			r.logger.Warn("Failed to unmarshal config tags", zap.Error(err))
			config.Tags = []string{}
		}
		if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal config metadata", zap.Error(err))
			config.Metadata = make(map[string]interface{})
		}

		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating configurations: %w", err)
	}

	return configs, total, nil
}

// DeleteConfig deletes a configuration by ID
func (r *DatabaseRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM configurations WHERE id = $1"
	
	result, err := r.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete configuration", zap.Error(err))
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("configuration not found: %s", id.String())
	}

	r.logger.Info("Configuration deleted successfully", zap.String("id", id.String()))
	return nil
}

// GetLatestConfig retrieves the latest configuration for a service and environment
func (r *DatabaseRepository) GetLatestConfig(ctx context.Context, service, environment string) (*models.Configuration, error) {
	query := `
		SELECT id, service, environment, version, data, schema, tags, metadata,
			   created_by, updated_by, created_at, updated_at
		FROM configurations 
		WHERE service = $1 AND environment = $2
		ORDER BY version DESC, created_at DESC
		LIMIT 1`

	var config models.Configuration
	var dataJSON, schemaJSON, tagsJSON, metadataJSON []byte

	err := r.db.DB().QueryRowContext(ctx, query, service, environment).Scan(
		&config.ID, &config.Service, &config.Environment, &config.Version,
		&dataJSON, &schemaJSON, &tagsJSON, &metadataJSON,
		&config.CreatedBy, &config.UpdatedBy, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("configuration not found for service: %s, environment: %s", service, environment)
		}
		r.logger.Error("Failed to get latest configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to get latest configuration: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(dataJSON, &config.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}
	if err := json.Unmarshal(schemaJSON, &config.Schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config schema: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &config.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config tags: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config metadata: %w", err)
	}

	return &config, nil
}

// GetConfigHistory retrieves configuration history for a specific config
func (r *DatabaseRepository) GetConfigHistory(ctx context.Context, configID uuid.UUID, page, limit int) ([]*models.Configuration, int64, error) {
	// First get the service and environment from the config
	var service, environment string
	err := r.db.DB().QueryRowContext(ctx, 
		"SELECT service, environment FROM configurations WHERE id = $1", 
		configID).Scan(&service, &environment)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get config details: %w", err)
	}

	// Now get all configurations for this service and environment
	return r.GetConfigs(ctx, page, limit, service, environment)
}

// CreateTemplate creates a new configuration template
func (r *DatabaseRepository) CreateTemplate(ctx context.Context, template *models.ConfigTemplate) error {
	query := `
		INSERT INTO config_templates (
			id, name, description, category, schema, defaults, tags, metadata,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	schemaJSON, err := json.Marshal(template.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal template schema: %w", err)
	}

	defaultsJSON, err := json.Marshal(template.Defaults)
	if err != nil {
		return fmt.Errorf("failed to marshal template defaults: %w", err)
	}

	tagsJSON, err := json.Marshal(template.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal template tags: %w", err)
	}

	metadataJSON, err := json.Marshal(template.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal template metadata: %w", err)
	}

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	_, err = r.db.DB().ExecContext(ctx, query,
		template.ID, template.Name, template.Description, template.Category,
		schemaJSON, defaultsJSON, tagsJSON, metadataJSON,
		template.CreatedBy, template.UpdatedBy, template.CreatedAt, template.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create template", zap.Error(err))
		return fmt.Errorf("failed to create template: %w", err)
	}

	r.logger.Info("Template created successfully", zap.String("id", template.ID.String()))
	return nil
}

// GetTemplates retrieves templates with pagination and filtering
func (r *DatabaseRepository) GetTemplates(ctx context.Context, page, limit int, category string) ([]*models.ConfigTemplate, int64, error) {
	offset := (page - 1) * limit
	
	// Build query with optional filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM config_templates %s", whereClause)
	var total int64
	err := r.db.DB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, name, description, category, schema, defaults, tags, metadata,
			   created_by, updated_by, created_at, updated_at
		FROM config_templates %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)

	rows, err := r.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.ConfigTemplate
	for rows.Next() {
		var template models.ConfigTemplate
		var schemaJSON, defaultsJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.Category,
			&schemaJSON, &defaultsJSON, &tagsJSON, &metadataJSON,
			&template.CreatedBy, &template.UpdatedBy, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan template: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(schemaJSON, &template.Schema); err != nil {
			r.logger.Warn("Failed to unmarshal template schema", zap.Error(err))
			template.Schema = make(map[string]interface{})
		}
		if err := json.Unmarshal(defaultsJSON, &template.Defaults); err != nil {
			r.logger.Warn("Failed to unmarshal template defaults", zap.Error(err))
			template.Defaults = make(map[string]interface{})
		}
		if err := json.Unmarshal(tagsJSON, &template.Tags); err != nil {
			r.logger.Warn("Failed to unmarshal template tags", zap.Error(err))
			template.Tags = []string{}
		}
		if err := json.Unmarshal(metadataJSON, &template.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal template metadata", zap.Error(err))
			template.Metadata = make(map[string]interface{})
		}

		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, total, nil
}

// GetTemplate retrieves a template by ID
func (r *DatabaseRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*models.ConfigTemplate, error) {
	query := `
		SELECT id, name, description, category, schema, defaults, tags, metadata,
			   created_by, updated_by, created_at, updated_at
		FROM config_templates 
		WHERE id = $1`

	var template models.ConfigTemplate
	var schemaJSON, defaultsJSON, tagsJSON, metadataJSON []byte

	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(
		&template.ID, &template.Name, &template.Description, &template.Category,
		&schemaJSON, &defaultsJSON, &tagsJSON, &metadataJSON,
		&template.CreatedBy, &template.UpdatedBy, &template.CreatedAt, &template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", id.String())
		}
		r.logger.Error("Failed to get template", zap.Error(err))
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(schemaJSON, &template.Schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template schema: %w", err)
	}
	if err := json.Unmarshal(defaultsJSON, &template.Defaults); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template defaults: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template tags: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &template.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template metadata: %w", err)
	}

	return &template, nil
}