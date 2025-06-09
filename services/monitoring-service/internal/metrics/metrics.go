package metrics

import (
    "context"
    "go.uber.org/zap"
)

// MetricsCollector collects system metrics
type MetricsCollector struct {
    logger *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
    return &MetricsCollector{logger: logger}
}

// Collect collects metrics
func (m *MetricsCollector) Collect(ctx context.Context) error {
    m.logger.Info("Collecting metrics")
    return nil
}
