package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"notification-service/internal/config"
	"go.uber.org/zap"
)

// emailProvider implements the EmailProvider interface
type emailProvider struct {
	config *config.EmailConfig
	logger *zap.Logger
}

// NewEmailProvider creates a new email provider
func NewEmailProvider(cfg config.EmailConfig, logger *zap.Logger) (EmailProvider, error) {
	provider := &emailProvider{
		config: &cfg,
		logger: logger,
	}

	// Validate configuration based on provider type
	if err := provider.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid email configuration: %w", err)
	}

	return provider, nil
}

// SendEmail sends a single email
func (p *emailProvider) SendEmail(ctx context.Context, req *EmailRequest) (*EmailResponse, error) {
	p.logger.Info("Sending email",
		zap.String("id", req.ID),
		zap.String("to", strings.Join(req.To, ",")),
		zap.String("subject", req.Subject),
		zap.String("provider", p.config.Provider),
	)

	response := &EmailResponse{
		ID:       req.ID,
		Provider: p.config.Provider,
		SentAt:   time.Now(),
		Status:   "sent",
	}

	switch p.config.Provider {
	case "smtp":
		return p.sendSMTPEmail(ctx, req, response)
	case "sendgrid":
		return p.sendSendGridEmail(ctx, req, response)
	case "ses":
		return p.sendSESEmail(ctx, req, response)
	case "mailgun":
		return p.sendMailgunEmail(ctx, req, response)
	default:
		return p.sendMockEmail(ctx, req, response)
	}
}

// SendBulkEmails sends multiple emails
func (p *emailProvider) SendBulkEmails(ctx context.Context, requests []*EmailRequest) ([]*EmailResponse, error) {
	p.logger.Info("Sending bulk emails",
		zap.Int("count", len(requests)),
		zap.String("provider", p.config.Provider),
	)

	responses := make([]*EmailResponse, len(requests))
	for i, req := range requests {
		resp, err := p.SendEmail(ctx, req)
		if err != nil {
			resp = &EmailResponse{
				ID:       req.ID,
				Provider: p.config.Provider,
				SentAt:   time.Now(),
				Status:   "failed",
				Error:    err.Error(),
			}
		}
		responses[i] = resp
	}

	return responses, nil
}

// HealthCheck checks if the email provider is healthy
func (p *emailProvider) HealthCheck() error {
	switch p.config.Provider {
	case "smtp":
		return p.healthCheckSMTP()
	case "sendgrid":
		return p.healthCheckSendGrid()
	case "ses":
		return p.healthCheckSES()
	case "mailgun":
		return p.healthCheckMailgun()
	default:
		return nil // Mock provider is always healthy
	}
}

// GetProviderName returns the provider name
func (p *emailProvider) GetProviderName() string {
	return p.config.Provider
}

// GetRateLimit returns the rate limit
func (p *emailProvider) GetRateLimit() int {
	return p.config.RateLimit
}

// validateConfig validates the email configuration
func (p *emailProvider) validateConfig() error {
	if !p.config.Enabled {
		return nil
	}

	switch p.config.Provider {
	case "smtp":
		if p.config.SMTP.Host == "" || p.config.SMTP.Username == "" || p.config.SMTP.Password == "" {
			return fmt.Errorf("SMTP configuration incomplete")
		}
	case "sendgrid":
		if p.config.SendGrid.APIKey == "" {
			return fmt.Errorf("SendGrid API key required")
		}
	case "ses":
		if p.config.SES.AccessKeyID == "" || p.config.SES.SecretAccessKey == "" {
			return fmt.Errorf("AWS SES credentials required")
		}
	case "mailgun":
		if p.config.Mailgun.Domain == "" || p.config.Mailgun.APIKey == "" {
			return fmt.Errorf("Mailgun configuration incomplete")
		}
	}

	return nil
}

// sendSMTPEmail sends email via SMTP
func (p *emailProvider) sendSMTPEmail(ctx context.Context, req *EmailRequest, response *EmailResponse) (*EmailResponse, error) {
	// Mock SMTP implementation for now
	p.logger.Info("Sending SMTP email (mock)",
		zap.String("host", p.config.SMTP.Host),
		zap.Int("port", p.config.SMTP.Port),
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	response.MessageID = fmt.Sprintf("smtp-%d", time.Now().Unix())
	return response, nil
}

// sendSendGridEmail sends email via SendGrid
func (p *emailProvider) sendSendGridEmail(ctx context.Context, req *EmailRequest, response *EmailResponse) (*EmailResponse, error) {
	// Mock SendGrid implementation for now
	p.logger.Info("Sending SendGrid email (mock)",
		zap.String("api_key", p.config.SendGrid.APIKey[:10]+"..."),
	)

	// Simulate processing time
	time.Sleep(150 * time.Millisecond)

	response.MessageID = fmt.Sprintf("sendgrid-%d", time.Now().Unix())
	return response, nil
}

// sendSESEmail sends email via AWS SES
func (p *emailProvider) sendSESEmail(ctx context.Context, req *EmailRequest, response *EmailResponse) (*EmailResponse, error) {
	// Mock SES implementation for now
	p.logger.Info("Sending SES email (mock)",
		zap.String("region", p.config.SES.Region),
	)

	// Simulate processing time
	time.Sleep(120 * time.Millisecond)

	response.MessageID = fmt.Sprintf("ses-%d", time.Now().Unix())
	return response, nil
}

// sendMailgunEmail sends email via Mailgun
func (p *emailProvider) sendMailgunEmail(ctx context.Context, req *EmailRequest, response *EmailResponse) (*EmailResponse, error) {
	// Mock Mailgun implementation for now
	p.logger.Info("Sending Mailgun email (mock)",
		zap.String("domain", p.config.Mailgun.Domain),
	)

	// Simulate processing time
	time.Sleep(130 * time.Millisecond)

	response.MessageID = fmt.Sprintf("mailgun-%d", time.Now().Unix())
	return response, nil
}

// sendMockEmail sends a mock email (for testing)
func (p *emailProvider) sendMockEmail(ctx context.Context, req *EmailRequest, response *EmailResponse) (*EmailResponse, error) {
	p.logger.Info("Sending mock email",
		zap.String("id", req.ID),
		zap.String("subject", req.Subject),
	)

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	response.MessageID = fmt.Sprintf("mock-%d", time.Now().Unix())
	return response, nil
}

// healthCheckSMTP checks SMTP connectivity
func (p *emailProvider) healthCheckSMTP() error {
	// Mock health check for now
	p.logger.Debug("SMTP health check (mock)")
	return nil
}

// healthCheckSendGrid checks SendGrid connectivity
func (p *emailProvider) healthCheckSendGrid() error {
	// Mock health check for now
	p.logger.Debug("SendGrid health check (mock)")
	return nil
}

// healthCheckSES checks AWS SES connectivity
func (p *emailProvider) healthCheckSES() error {
	// Mock health check for now
	p.logger.Debug("SES health check (mock)")
	return nil
}

// healthCheckMailgun checks Mailgun connectivity
func (p *emailProvider) healthCheckMailgun() error {
	// Mock health check for now
	p.logger.Debug("Mailgun health check (mock)")
	return nil
}

// Helper function to create SMTP auth
func (p *emailProvider) createSMTPAuth() smtp.Auth {
	return smtp.PlainAuth("", p.config.SMTP.Username, p.config.SMTP.Password, p.config.SMTP.Host)
}

// Helper function to create TLS config
func (p *emailProvider) createTLSConfig() *tls.Config {
	return &tls.Config{
		ServerName: p.config.SMTP.Host,
	}
}