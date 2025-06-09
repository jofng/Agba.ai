package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/agba-ai/freeswitch-service/internal/config"
	"github.com/agba-ai/freeswitch-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ServiceIntegration handles integration with other services
type ServiceIntegration struct {
	config *config.Config
	logger *zap.Logger
	client *http.Client
}

// NewServiceIntegration creates a new service integration
func NewServiceIntegration(config *config.Config, logger *zap.Logger) *ServiceIntegration {
	return &ServiceIntegration{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StoreCall stores a call record in the database
func (s *ServiceIntegration) StoreCall(ctx context.Context, call *models.Call) error {
	// This would typically store in a local database
	// For now, we'll log the call and potentially send to monitoring service
	s.logger.Info("Storing call",
		zap.String("call_id", call.ID.String()),
		zap.String("session_id", call.SessionID),
		zap.String("caller_id", call.CallerID),
		zap.String("called_number", call.CalledNumber))

	// Send call metrics to monitoring service
	return s.SendCallMetrics(ctx, call, "call_stored")
}

// GetCall retrieves a call by ID
func (s *ServiceIntegration) GetCall(ctx context.Context, callID uuid.UUID) (*models.Call, error) {
	// This would typically query a local database
	// For now, return a mock call
	s.logger.Info("Getting call", zap.String("call_id", callID.String()))

	// In a real implementation, this would query the database
	return &models.Call{
		ID:        callID,
		SessionID: "mock-session-" + callID.String(),
		Status:    models.CallStatusAnswered,
	}, nil
}

// GetCallBySessionID retrieves a call by session ID
func (s *ServiceIntegration) GetCallBySessionID(ctx context.Context, sessionID string) (*models.Call, error) {
	s.logger.Info("Getting call by session ID", zap.String("session_id", sessionID))

	// In a real implementation, this would query the database
	return &models.Call{
		ID:        uuid.New(),
		SessionID: sessionID,
		Status:    models.CallStatusAnswered,
	}, nil
}

// UpdateCall updates a call record
func (s *ServiceIntegration) UpdateCall(ctx context.Context, call *models.Call) error {
	s.logger.Info("Updating call",
		zap.String("call_id", call.ID.String()),
		zap.String("status", string(call.Status)))

	// Send updated metrics
	return s.SendCallMetrics(ctx, call, "call_updated")
}

// GetCallStats retrieves call statistics
func (s *ServiceIntegration) GetCallStats(ctx context.Context) (*models.CallStats, error) {
	s.logger.Info("Getting call statistics")

	// In a real implementation, this would query the database for stats
	stats := &models.CallStats{
		TotalCalls:      1000,
		ActiveCalls:     25,
		SuccessfulCalls: 950,
		FailedCalls:     50,
		AverageCallTime: 180.5,
		CallSuccessRate: 95.0,
		PeakConcurrency: 100,
		LastUpdated:     time.Now(),
	}

	return stats, nil
}

// StartRecording starts a recording for a call
func (s *ServiceIntegration) StartRecording(ctx context.Context, call *models.Call) (uuid.UUID, string, error) {
	s.logger.Info("Starting recording for call", zap.String("call_id", call.ID.String()))

	recordingID := uuid.New()
	filename := fmt.Sprintf("recording_%s_%d.wav", call.SessionID, time.Now().Unix())

	// Call recording service to start recording
	recordingRequest := map[string]interface{}{
		"session_id": call.SessionID,
		"user_id":    call.UserID,
		"filename":   filename,
		"format":     "wav",
		"quality":    "high",
	}

	if err := s.callRecordingService(ctx, "POST", "/recordings", recordingRequest); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to start recording: %w", err)
	}

	return recordingID, filename, nil
}

// StopRecording stops a recording
func (s *ServiceIntegration) StopRecording(ctx context.Context, recordingID uuid.UUID) error {
	s.logger.Info("Stopping recording", zap.String("recording_id", recordingID.String()))

	// Call recording service to stop recording
	return s.callRecordingService(ctx, "PUT", fmt.Sprintf("/recordings/%s/stop", recordingID), nil)
}

// CompleteRecording marks a recording as complete
func (s *ServiceIntegration) CompleteRecording(ctx context.Context, recordingID uuid.UUID) error {
	s.logger.Info("Completing recording", zap.String("recording_id", recordingID.String()))

	// Call recording service to complete recording
	return s.callRecordingService(ctx, "PUT", fmt.Sprintf("/recordings/%s/complete", recordingID), nil)
}

// GetRecordingFilename gets the filename for a recording
func (s *ServiceIntegration) GetRecordingFilename(ctx context.Context, recordingID uuid.UUID) (string, error) {
	s.logger.Info("Getting recording filename", zap.String("recording_id", recordingID.String()))

	// In a real implementation, this would query the recording service
	return fmt.Sprintf("recording_%s.wav", recordingID.String()), nil
}

// SendCallMetrics sends call metrics to the monitoring service
func (s *ServiceIntegration) SendCallMetrics(ctx context.Context, call *models.Call, eventType string) error {
	metrics := map[string]interface{}{
		"service_name": "freeswitch-service",
		"metric_name":  "call_event",
		"metric_value": 1,
		"metric_unit":  "count",
		"labels": map[string]string{
			"event_type":    eventType,
			"call_id":       call.ID.String(),
			"session_id":    call.SessionID,
			"direction":     string(call.Direction),
			"status":        string(call.Status),
			"caller_id":     call.CallerID,
			"called_number": call.CalledNumber,
		},
		"timestamp": time.Now(),
	}

	return s.callMonitoringService(ctx, "POST", "/metrics", metrics)
}

// SendNotification sends a notification through the notification service
func (s *ServiceIntegration) SendNotification(ctx context.Context, notification map[string]interface{}) error {
	s.logger.Info("Sending notification")

	return s.callNotificationService(ctx, "POST", "/notifications", notification)
}

// HealthCheck checks the health of integrated services
func (s *ServiceIntegration) HealthCheck(ctx context.Context) map[string]interface{} {
	checks := make(map[string]interface{})

	// Check recording service
	if err := s.callRecordingService(ctx, "GET", "/health", nil); err != nil {
		checks["recording_service"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	} else {
		checks["recording_service"] = map[string]interface{}{
			"status":  "healthy",
			"message": "Recording service is accessible",
		}
	}

	// Check monitoring service
	if err := s.callMonitoringService(ctx, "GET", "/health", nil); err != nil {
		checks["monitoring_service"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	} else {
		checks["monitoring_service"] = map[string]interface{}{
			"status":  "healthy",
			"message": "Monitoring service is accessible",
		}
	}

	// Check user service
	if err := s.callUserService(ctx, "GET", "/health", nil); err != nil {
		checks["user_service"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	} else {
		checks["user_service"] = map[string]interface{}{
			"status":  "healthy",
			"message": "User service is accessible",
		}
	}

	// Check config service
	if err := s.callConfigService(ctx, "GET", "/health", nil); err != nil {
		checks["config_service"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	} else {
		checks["config_service"] = map[string]interface{}{
			"status":  "healthy",
			"message": "Config service is accessible",
		}
	}

	// Check notification service
	if err := s.callNotificationService(ctx, "GET", "/health", nil); err != nil {
		checks["notification_service"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	} else {
		checks["notification_service"] = map[string]interface{}{
			"status":  "healthy",
			"message": "Notification service is accessible",
		}
	}

	return checks
}

// Helper methods for calling other services
func (s *ServiceIntegration) callRecordingService(ctx context.Context, method, path string, data interface{}) error {
	url := fmt.Sprintf("http://%s:%d%s",
		s.config.Services.RecordingService.Host,
		s.config.Services.RecordingService.Port,
		path)

	return s.makeHTTPRequest(ctx, method, url, data)
}

func (s *ServiceIntegration) callMonitoringService(ctx context.Context, method, path string, data interface{}) error {
	url := fmt.Sprintf("http://%s:%d%s",
		s.config.Services.MonitoringService.Host,
		s.config.Services.MonitoringService.Port,
		path)

	return s.makeHTTPRequest(ctx, method, url, data)
}

func (s *ServiceIntegration) callUserService(ctx context.Context, method, path string, data interface{}) error {
	url := fmt.Sprintf("http://%s:%d%s",
		s.config.Services.UserService.Host,
		s.config.Services.UserService.Port,
		path)

	return s.makeHTTPRequest(ctx, method, url, data)
}

func (s *ServiceIntegration) callConfigService(ctx context.Context, method, path string, data interface{}) error {
	url := fmt.Sprintf("http://%s:%d%s",
		s.config.Services.ConfigService.Host,
		s.config.Services.ConfigService.Port,
		path)

	return s.makeHTTPRequest(ctx, method, url, data)
}

func (s *ServiceIntegration) callNotificationService(ctx context.Context, method, path string, data interface{}) error {
	url := fmt.Sprintf("http://%s:%d%s",
		s.config.Services.NotificationService.Host,
		s.config.Services.NotificationService.Port,
		path)

	return s.makeHTTPRequest(ctx, method, url, data)
}

func (s *ServiceIntegration) makeHTTPRequest(ctx context.Context, method, url string, data interface{}) error {
	var body []byte
	var err error

	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal request data: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return nil
}