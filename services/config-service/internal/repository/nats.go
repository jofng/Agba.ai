package repository

import (
	"fmt"
	"time"

	"config-service/internal/config"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NATS represents the NATS connection
type NATS struct {
	conn   *nats.Conn
	logger *zap.Logger
}

// NewNATS creates a new NATS connection
func NewNATS(cfg config.NATSConfig, logger *zap.Logger) (*NATS, error) {
	opts := []nats.Option{
		nats.UserInfo(cfg.Username, cfg.Password),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected")
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("NATS connection established",
		zap.String("url", cfg.URL),
		zap.String("username", cfg.Username),
	)

	return &NATS{
		conn:   nc,
		logger: logger,
	}, nil
}

// Conn returns the underlying NATS connection
func (n *NATS) Conn() *nats.Conn {
	return n.conn
}

// Close closes the NATS connection
func (n *NATS) Close() error {
	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

// IsConnected returns whether NATS is connected
func (n *NATS) IsConnected() bool {
	return n.conn != nil && n.conn.IsConnected()
}

// Health returns the health status of NATS
func (n *NATS) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"connected": n.IsConnected(),
	}

	if !n.IsConnected() {
		health["status"] = "unhealthy"
		health["error"] = "not connected"
	}

	if n.conn != nil {
		stats := n.conn.Statistics
		health["stats"] = map[string]interface{}{
			"in_msgs":     stats.InMsgs,
			"out_msgs":    stats.OutMsgs,
			"in_bytes":    stats.InBytes,
			"out_bytes":   stats.OutBytes,
			"reconnects":  stats.Reconnects,
		}
	}

	return health
}