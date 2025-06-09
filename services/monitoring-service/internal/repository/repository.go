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

// HealthCheck checks repository health
func (r *Repository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
