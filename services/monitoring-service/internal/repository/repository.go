package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/agba-ai/monitoring-service/internal/models"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

// Repository handles data operations for monitoring service
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

// StoreMetric stores a metric in the database
func (r *Repository) StoreMetric(ctx context.Context, metric *models.Metric) error {
    query := `
        INSERT INTO metrics (id, service_name, metric_name, metric_value, metric_unit, labels, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`
    
    labelsJSON, err := json.Marshal(metric.Labels)
    if err != nil {
        return fmt.Errorf("failed to marshal labels: %w", err)
    }

    _, err = r.db.ExecContext(ctx, query,
        uuid.New(), metric.ServiceName, metric.Name, metric.Value,
        metric.Unit, labelsJSON, metric.Timestamp)
    
    if err != nil {
        r.logger.Error("Failed to store metric", zap.Error(err))
        return fmt.Errorf("failed to store metric: %w", err)
    }

    return nil
}

// GetMetrics retrieves metrics from the database
func (r *Repository) GetMetrics(ctx context.Context, serviceName, metricName string, since time.Time) ([]*models.Metric, error) {
    query := `
        SELECT service_name, metric_name, metric_value, metric_unit, labels, timestamp
        FROM metrics
        WHERE service_name = $1 AND metric_name = $2 AND timestamp >= $3
        ORDER BY timestamp DESC`

    rows, err := r.db.QueryContext(ctx, query, serviceName, metricName, since)
    if err != nil {
        return nil, fmt.Errorf("failed to query metrics: %w", err)
    }
    defer rows.Close()

    var metrics []*models.Metric
    for rows.Next() {
        var metric models.Metric
        var labelsJSON []byte

        err := rows.Scan(
            &metric.ServiceName, &metric.Name, &metric.Value,
            &metric.Unit, &labelsJSON, &metric.Timestamp)
        if err != nil {
            return nil, fmt.Errorf("failed to scan metric: %w", err)
        }

        if err := json.Unmarshal(labelsJSON, &metric.Labels); err != nil {
            r.logger.Warn("Failed to unmarshal labels", zap.Error(err))
            metric.Labels = make(map[string]string)
        }

        metrics = append(metrics, &metric)
    }

    return metrics, nil
}

// CreateAlert creates a new alert rule
func (r *Repository) CreateAlert(ctx context.Context, alert *models.Alert) error {
    query := `
        INSERT INTO alerts (id, name, description, service, metric, condition, threshold, severity, enabled, created_by, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

    now := time.Now()
    _, err := r.db.ExecContext(ctx, query,
        alert.ID, alert.Name, alert.Description, alert.Service,
        alert.Metric, alert.Condition, alert.Threshold, alert.Severity,
        alert.Enabled, alert.CreatedBy, now, now)

    if err != nil {
        r.logger.Error("Failed to create alert", zap.Error(err))
        return fmt.Errorf("failed to create alert: %w", err)
    }

    return nil
}

// GetAlerts retrieves alerts with pagination
func (r *Repository) GetAlerts(ctx context.Context, page, limit int, service, status string) ([]*models.Alert, int64, error) {
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

    // Count query
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts %s", whereClause)
    var total int64
    err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
    }

    // Data query
    query := fmt.Sprintf(`
        SELECT id, name, description, service, metric, condition, threshold, severity, enabled, created_by, created_at, updated_at
        FROM alerts %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
    
    args = append(args, limit, offset)

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to query alerts: %w", err)
    }
    defer rows.Close()

    var alerts []*models.Alert
    for rows.Next() {
        var alert models.Alert
        err := rows.Scan(
            &alert.ID, &alert.Name, &alert.Description, &alert.Service,
            &alert.Metric, &alert.Condition, &alert.Threshold, &alert.Severity,
            &alert.Enabled, &alert.CreatedBy, &alert.CreatedAt, &alert.UpdatedAt)
        if err != nil {
            return nil, 0, fmt.Errorf("failed to scan alert: %w", err)
        }
        alerts = append(alerts, &alert)
    }

    return alerts, total, nil
}

// UpdateAlert updates an existing alert rule
func (r *Repository) UpdateAlert(ctx context.Context, alert *models.Alert) error {
	query := `
		UPDATE alerts 
		SET name = $2, description = $3, service = $4, metric = $5, condition = $6, 
		    threshold = $7, severity = $8, enabled = $9, updated_by = $10, updated_at = $11
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.Name, alert.Description, alert.Service,
		alert.Metric, alert.Condition, alert.Threshold, alert.Severity,
		alert.Enabled, alert.UpdatedBy, time.Now())

	if err != nil {
		r.logger.Error("Failed to update alert", zap.Error(err))
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

// DeleteAlert deletes an alert rule
func (r *Repository) DeleteAlert(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM alerts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete alert", zap.Error(err))
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}

// GetAlert retrieves a single alert by ID
func (r *Repository) GetAlert(ctx context.Context, id uuid.UUID) (*models.Alert, error) {
	query := `
		SELECT id, name, description, service, metric, condition, threshold, severity, enabled, created_by, created_at, updated_at
		FROM alerts
		WHERE id = $1`

	var alert models.Alert
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.Name, &alert.Description, &alert.Service,
		&alert.Metric, &alert.Condition, &alert.Threshold, &alert.Severity,
		&alert.Enabled, &alert.CreatedBy, &alert.CreatedAt, &alert.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return &alert, nil
}

// StoreAlertEvent stores an alert event
func (r *Repository) StoreAlertEvent(ctx context.Context, event *models.AlertEvent) error {
	query := `
		INSERT INTO alert_events (id, alert_id, type, message, value, threshold, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.AlertID, event.Type, event.Message,
		event.Value, event.Threshold, event.Timestamp)

	if err != nil {
		r.logger.Error("Failed to store alert event", zap.Error(err))
		return fmt.Errorf("failed to store alert event: %w", err)
	}

	return nil
}

// GetAlertEvents retrieves alert events with pagination
func (r *Repository) GetAlertEvents(ctx context.Context, alertID uuid.UUID, page, limit int) ([]*models.AlertEvent, int64, error) {
	offset := (page - 1) * limit

	// Count query
	countQuery := `SELECT COUNT(*) FROM alert_events WHERE alert_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, alertID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alert events: %w", err)
	}

	// Data query
	query := `
		SELECT id, alert_id, type, message, value, threshold, timestamp
		FROM alert_events
		WHERE alert_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, alertID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query alert events: %w", err)
	}
	defer rows.Close()

	var events []*models.AlertEvent
	for rows.Next() {
		var event models.AlertEvent
		err := rows.Scan(
			&event.ID, &event.AlertID, &event.Type, &event.Message,
			&event.Value, &event.Threshold, &event.Timestamp)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan alert event: %w", err)
		}
		events = append(events, &event)
	}

	return events, total, nil
}

// StoreLogEntry stores a log entry
func (r *Repository) StoreLogEntry(ctx context.Context, entry *models.LogEntry) error {
	query := `
		INSERT INTO log_entries (id, service, level, message, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)`

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		entry.ID, entry.Service, entry.Level, entry.Message,
		metadataJSON, entry.Timestamp)

	if err != nil {
		r.logger.Error("Failed to store log entry", zap.Error(err))
		return fmt.Errorf("failed to store log entry: %w", err)
	}

	return nil
}

// GetLogEntries retrieves log entries with pagination and filtering
func (r *Repository) GetLogEntries(ctx context.Context, service, level string, since time.Time, page, limit int) ([]*models.LogEntry, int64, error) {
	offset := (page - 1) * limit

	// Build query with optional filters
	whereClause := "WHERE timestamp >= $1"
	args := []interface{}{since}
	argIndex := 2

	if service != "" {
		whereClause += fmt.Sprintf(" AND service = $%d", argIndex)
		args = append(args, service)
		argIndex++
	}

	if level != "" {
		whereClause += fmt.Sprintf(" AND level = $%d", argIndex)
		args = append(args, level)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM log_entries %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count log entries: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, service, level, message, metadata, timestamp
		FROM log_entries %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query log entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
		var metadataJSON []byte

		err := rows.Scan(
			&entry.ID, &entry.Service, &entry.Level, &entry.Message,
			&metadataJSON, &entry.Timestamp)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log entry: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &entry.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal metadata", zap.Error(err))
			entry.Metadata = make(map[string]interface{})
		}

		entries = append(entries, &entry)
	}

	return entries, total, nil
}



// HealthCheck checks repository health
func (r *Repository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
