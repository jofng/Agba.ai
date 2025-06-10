package providers

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/config"
	"go.uber.org/zap"
)

// smsProvider implements the SMSProvider interface
type smsProvider struct {
	config *config.SMSConfig
	logger *zap.Logger
}

// NewSMSProvider creates a new SMS provider
func NewSMSProvider(cfg config.SMSConfig, logger *zap.Logger) (SMSProvider, error) {
	provider := &smsProvider{
		config: &cfg,
		logger: logger,
	}

	// Validate configuration based on provider type
	if err := provider.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid SMS configuration: %w", err)
	}

	return provider, nil
}

// SendSMS sends a single SMS
func (p *smsProvider) SendSMS(ctx context.Context, req *SMSRequest) (*SMSResponse, error) {
	p.logger.Info("Sending SMS",
		zap.String("id", req.ID),
		zap.String("to", req.To),
		zap.String("from", req.From),
		zap.String("provider", p.config.Provider),
	)

	response := &SMSResponse{
		ID:       req.ID,
		Provider: p.config.Provider,
		SentAt:   time.Now(),
		Status:   "sent",
		Segments: p.calculateSegments(req.Body),
	}

	switch p.config.Provider {
	case "twilio":
		return p.sendTwilioSMS(ctx, req, response)
	case "aws_sns":
		return p.sendSNSSMS(ctx, req, response)
	case "nexmo":
		return p.sendNexmoSMS(ctx, req, response)
	default:
		return p.sendMockSMS(ctx, req, response)
	}
}

// SendBulkSMS sends multiple SMS messages
func (p *smsProvider) SendBulkSMS(ctx context.Context, requests []*SMSRequest) ([]*SMSResponse, error) {
	p.logger.Info("Sending bulk SMS",
		zap.Int("count", len(requests)),
		zap.String("provider", p.config.Provider),
	)

	responses := make([]*SMSResponse, len(requests))
	for i, req := range requests {
		resp, err := p.SendSMS(ctx, req)
		if err != nil {
			resp = &SMSResponse{
				ID:       req.ID,
				Provider: p.config.Provider,
				SentAt:   time.Now(),
				Status:   "failed",
				Error:    err.Error(),
				Segments: p.calculateSegments(req.Body),
			}
		}
		responses[i] = resp
	}

	return responses, nil
}

// HealthCheck checks if the SMS provider is healthy
func (p *smsProvider) HealthCheck() error {
	switch p.config.Provider {
	case "twilio":
		return p.healthCheckTwilio()
	case "aws_sns":
		return p.healthCheckSNS()
	case "nexmo":
		return p.healthCheckNexmo()
	default:
		return nil // Mock provider is always healthy
	}
}

// GetProviderName returns the provider name
func (p *smsProvider) GetProviderName() string {
	return p.config.Provider
}

// GetRateLimit returns the rate limit
func (p *smsProvider) GetRateLimit() int {
	return p.config.RateLimit
}

// validateConfig validates the SMS configuration
func (p *smsProvider) validateConfig() error {
	if !p.config.Enabled {
		return nil
	}

	switch p.config.Provider {
	case "twilio":
		if p.config.Twilio.AccountSID == "" || p.config.Twilio.AuthToken == "" {
			return fmt.Errorf("Twilio configuration incomplete")
		}
	case "aws_sns":
		if p.config.SNS.AccessKeyID == "" || p.config.SNS.SecretAccessKey == "" {
			return fmt.Errorf("AWS SNS credentials required")
		}
	case "nexmo":
		if p.config.Nexmo.APIKey == "" || p.config.Nexmo.APISecret == "" {
			return fmt.Errorf("Nexmo configuration incomplete")
		}
	}

	return nil
}

// sendTwilioSMS sends SMS via Twilio
func (p *smsProvider) sendTwilioSMS(ctx context.Context, req *SMSRequest, response *SMSResponse) (*SMSResponse, error) {
	// Mock Twilio implementation for now
	p.logger.Info("Sending Twilio SMS (mock)",
		zap.String("account_sid", p.config.Twilio.AccountSID[:10]+"..."),
	)

	// Simulate processing time
	time.Sleep(200 * time.Millisecond)

	response.MessageID = fmt.Sprintf("twilio-%d", time.Now().Unix())
	response.Cost = 0.0075 * float64(response.Segments) // Mock cost calculation
	return response, nil
}

// sendSNSSMS sends SMS via AWS SNS
func (p *smsProvider) sendSNSSMS(ctx context.Context, req *SMSRequest, response *SMSResponse) (*SMSResponse, error) {
	// Mock SNS implementation for now
	p.logger.Info("Sending SNS SMS (mock)",
		zap.String("region", p.config.SNS.Region),
	)

	// Simulate processing time
	time.Sleep(180 * time.Millisecond)

	response.MessageID = fmt.Sprintf("sns-%d", time.Now().Unix())
	response.Cost = 0.0065 * float64(response.Segments) // Mock cost calculation
	return response, nil
}

// sendNexmoSMS sends SMS via Nexmo
func (p *smsProvider) sendNexmoSMS(ctx context.Context, req *SMSRequest, response *SMSResponse) (*SMSResponse, error) {
	// Mock Nexmo implementation for now
	p.logger.Info("Sending Nexmo SMS (mock)",
		zap.String("api_key", p.config.Nexmo.APIKey[:10]+"..."),
	)

	// Simulate processing time
	time.Sleep(190 * time.Millisecond)

	response.MessageID = fmt.Sprintf("nexmo-%d", time.Now().Unix())
	response.Cost = 0.0080 * float64(response.Segments) // Mock cost calculation
	return response, nil
}

// sendMockSMS sends a mock SMS (for testing)
func (p *smsProvider) sendMockSMS(ctx context.Context, req *SMSRequest, response *SMSResponse) (*SMSResponse, error) {
	p.logger.Info("Sending mock SMS",
		zap.String("id", req.ID),
		zap.String("to", req.To),
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	response.MessageID = fmt.Sprintf("mock-%d", time.Now().Unix())
	response.Cost = 0.01 * float64(response.Segments) // Mock cost calculation
	return response, nil
}

// healthCheckTwilio checks Twilio connectivity
func (p *smsProvider) healthCheckTwilio() error {
	// Mock health check for now
	p.logger.Debug("Twilio health check (mock)")
	return nil
}

// healthCheckSNS checks AWS SNS connectivity
func (p *smsProvider) healthCheckSNS() error {
	// Mock health check for now
	p.logger.Debug("SNS health check (mock)")
	return nil
}

// healthCheckNexmo checks Nexmo connectivity
func (p *smsProvider) healthCheckNexmo() error {
	// Mock health check for now
	p.logger.Debug("Nexmo health check (mock)")
	return nil
}

// calculateSegments calculates the number of SMS segments
func (p *smsProvider) calculateSegments(body string) int {
	maxLength := p.config.MaxLength
	if maxLength == 0 {
		maxLength = 160 // Default SMS length
	}

	// Simple calculation - in reality this would be more complex
	// considering encoding (GSM 7-bit vs UCS-2) and concatenation headers
	length := len(body)
	if length <= maxLength {
		return 1
	}

	// Account for concatenation headers (reduces usable space per segment)
	segmentLength := maxLength - 7 // Reserve space for concatenation headers
	return (length + segmentLength - 1) / segmentLength
}

// validatePhoneNumber validates a phone number format
func (p *smsProvider) validatePhoneNumber(phoneNumber string) error {
	// Basic validation - in reality this would be more comprehensive
	if len(phoneNumber) < 10 {
		return fmt.Errorf("phone number too short")
	}
	if len(phoneNumber) > 15 {
		return fmt.Errorf("phone number too long")
	}
	return nil
}

// formatPhoneNumber formats a phone number for the provider
func (p *smsProvider) formatPhoneNumber(phoneNumber string) string {
	// Basic formatting - in reality this would handle international formats
	if phoneNumber[0] != '+' {
		return "+" + phoneNumber
	}
	return phoneNumber
}