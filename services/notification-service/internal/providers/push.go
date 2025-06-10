package providers

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/config"
	"go.uber.org/zap"
)

// pushProvider implements the PushProvider interface
type pushProvider struct {
	config *config.PushConfig
	logger *zap.Logger
}

// NewPushProvider creates a new push notification provider
func NewPushProvider(cfg config.PushConfig, logger *zap.Logger) (PushProvider, error) {
	provider := &pushProvider{
		config: &cfg,
		logger: logger,
	}

	// Validate configuration based on provider type
	if err := provider.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid push configuration: %w", err)
	}

	return provider, nil
}

// SendPushNotification sends a single push notification
func (p *pushProvider) SendPushNotification(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	p.logger.Info("Sending push notification",
		zap.String("id", req.ID),
		zap.Int("device_count", len(req.DeviceTokens)),
		zap.String("title", req.Title),
		zap.String("provider", p.config.Provider),
		zap.String("platform", req.Platform),
	)

	response := &PushResponse{
		ID:           req.ID,
		Provider:     p.config.Provider,
		SentAt:       time.Now(),
		Status:       "sent",
		SuccessCount: len(req.DeviceTokens), // Assume all succeed for mock
		FailureCount: 0,
	}

	switch p.config.Provider {
	case "fcm":
		return p.sendFCMPush(ctx, req, response)
	case "apns":
		return p.sendAPNSPush(ctx, req, response)
	default:
		return p.sendMockPush(ctx, req, response)
	}
}

// SendBulkPushNotifications sends multiple push notifications
func (p *pushProvider) SendBulkPushNotifications(ctx context.Context, requests []*PushRequest) ([]*PushResponse, error) {
	p.logger.Info("Sending bulk push notifications",
		zap.Int("count", len(requests)),
		zap.String("provider", p.config.Provider),
	)

	responses := make([]*PushResponse, len(requests))
	for i, req := range requests {
		resp, err := p.SendPushNotification(ctx, req)
		if err != nil {
			resp = &PushResponse{
				ID:           req.ID,
				Provider:     p.config.Provider,
				SentAt:       time.Now(),
				Status:       "failed",
				SuccessCount: 0,
				FailureCount: len(req.DeviceTokens),
			}
		}
		responses[i] = resp
	}

	return responses, nil
}

// HealthCheck checks if the push provider is healthy
func (p *pushProvider) HealthCheck() error {
	switch p.config.Provider {
	case "fcm":
		return p.healthCheckFCM()
	case "apns":
		return p.healthCheckAPNS()
	default:
		return nil // Mock provider is always healthy
	}
}

// GetProviderName returns the provider name
func (p *pushProvider) GetProviderName() string {
	return p.config.Provider
}

// GetRateLimit returns the rate limit
func (p *pushProvider) GetRateLimit() int {
	return p.config.RateLimit
}

// validateConfig validates the push configuration
func (p *pushProvider) validateConfig() error {
	if !p.config.Enabled {
		return nil
	}

	switch p.config.Provider {
	case "fcm":
		if p.config.FCM.ServerKey == "" && p.config.FCM.Credentials == "" {
			return fmt.Errorf("FCM configuration incomplete - need server key or credentials")
		}
	case "apns":
		if p.config.APNS.KeyID == "" || p.config.APNS.TeamID == "" || p.config.APNS.BundleID == "" {
			return fmt.Errorf("APNS configuration incomplete")
		}
	}

	return nil
}

// sendFCMPush sends push notification via Firebase Cloud Messaging
func (p *pushProvider) sendFCMPush(ctx context.Context, req *PushRequest, response *PushResponse) (*PushResponse, error) {
	// Mock FCM implementation for now
	p.logger.Info("Sending FCM push notification (mock)",
		zap.String("project_id", p.config.FCM.ProjectID),
		zap.Int("device_count", len(req.DeviceTokens)),
	)

	// Simulate processing time based on device count
	processingTime := time.Duration(len(req.DeviceTokens)*10) * time.Millisecond
	if processingTime > 500*time.Millisecond {
		processingTime = 500 * time.Millisecond
	}
	time.Sleep(processingTime)

	response.MessageID = fmt.Sprintf("fcm-%d", time.Now().Unix())
	
	// Simulate some failures and invalid tokens
	if len(req.DeviceTokens) > 10 {
		response.FailureCount = len(req.DeviceTokens) / 20 // 5% failure rate
		response.SuccessCount = len(req.DeviceTokens) - response.FailureCount
		
		// Simulate some invalid tokens
		if response.FailureCount > 0 {
			response.InvalidTokens = []string{
				req.DeviceTokens[0] + "_invalid",
			}
		}
	}

	return response, nil
}

// sendAPNSPush sends push notification via Apple Push Notification Service
func (p *pushProvider) sendAPNSPush(ctx context.Context, req *PushRequest, response *PushResponse) (*PushResponse, error) {
	// Mock APNS implementation for now
	p.logger.Info("Sending APNS push notification (mock)",
		zap.String("bundle_id", p.config.APNS.BundleID),
		zap.Bool("production", p.config.APNS.Production),
		zap.Int("device_count", len(req.DeviceTokens)),
	)

	// Simulate processing time based on device count
	processingTime := time.Duration(len(req.DeviceTokens)*15) * time.Millisecond
	if processingTime > 600*time.Millisecond {
		processingTime = 600 * time.Millisecond
	}
	time.Sleep(processingTime)

	response.MessageID = fmt.Sprintf("apns-%d", time.Now().Unix())
	
	// Simulate some failures
	if len(req.DeviceTokens) > 5 {
		response.FailureCount = len(req.DeviceTokens) / 30 // ~3% failure rate
		response.SuccessCount = len(req.DeviceTokens) - response.FailureCount
	}

	return response, nil
}

// sendMockPush sends a mock push notification (for testing)
func (p *pushProvider) sendMockPush(ctx context.Context, req *PushRequest, response *PushResponse) (*PushResponse, error) {
	p.logger.Info("Sending mock push notification",
		zap.String("id", req.ID),
		zap.String("title", req.Title),
		zap.Int("device_count", len(req.DeviceTokens)),
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	response.MessageID = fmt.Sprintf("mock-%d", time.Now().Unix())
	return response, nil
}

// healthCheckFCM checks FCM connectivity
func (p *pushProvider) healthCheckFCM() error {
	// Mock health check for now
	p.logger.Debug("FCM health check (mock)")
	return nil
}

// healthCheckAPNS checks APNS connectivity
func (p *pushProvider) healthCheckAPNS() error {
	// Mock health check for now
	p.logger.Debug("APNS health check (mock)")
	return nil
}

// validateDeviceToken validates a device token format
func (p *pushProvider) validateDeviceToken(token string, platform string) error {
	if token == "" {
		return fmt.Errorf("device token cannot be empty")
	}

	switch platform {
	case "ios":
		// iOS tokens are typically 64 characters (hex)
		if len(token) != 64 {
			return fmt.Errorf("invalid iOS device token length")
		}
	case "android":
		// Android tokens are variable length but typically longer
		if len(token) < 100 {
			return fmt.Errorf("invalid Android device token length")
		}
	}

	return nil
}

// buildFCMPayload builds FCM payload from request
func (p *pushProvider) buildFCMPayload(req *PushRequest) map[string]interface{} {
	payload := map[string]interface{}{
		"registration_ids": req.DeviceTokens,
		"notification": map[string]interface{}{
			"title": req.Title,
			"body":  req.Body,
		},
	}

	if req.Icon != "" {
		payload["notification"].(map[string]interface{})["icon"] = req.Icon
	}

	if req.Sound != "" {
		payload["notification"].(map[string]interface{})["sound"] = req.Sound
	}

	if len(req.Data) > 0 {
		payload["data"] = req.Data
	}

	if req.TTL > 0 {
		payload["time_to_live"] = int(req.TTL.Seconds())
	}

	return payload
}

// buildAPNSPayload builds APNS payload from request
func (p *pushProvider) buildAPNSPayload(req *PushRequest) map[string]interface{} {
	aps := map[string]interface{}{
		"alert": map[string]interface{}{
			"title": req.Title,
			"body":  req.Body,
		},
	}

	if req.Badge > 0 {
		aps["badge"] = req.Badge
	}

	if req.Sound != "" {
		aps["sound"] = req.Sound
	}

	payload := map[string]interface{}{
		"aps": aps,
	}

	// Add custom data
	for key, value := range req.Data {
		payload[key] = value
	}

	return payload
}