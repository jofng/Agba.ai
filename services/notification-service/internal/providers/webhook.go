package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"notification-service/internal/config"
	"go.uber.org/zap"
)

// webhookProvider implements the WebhookProvider interface
type webhookProvider struct {
	config *config.WebhookConfig
	logger *zap.Logger
	client *http.Client
}

// NewWebhookProvider creates a new webhook provider
func NewWebhookProvider(cfg config.WebhookConfig, logger *zap.Logger) WebhookProvider {
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	return &webhookProvider{
		config: &cfg,
		logger: logger,
		client: client,
	}
}

// SendWebhook sends a single webhook
func (p *webhookProvider) SendWebhook(ctx context.Context, req *WebhookRequest) (*WebhookResponse, error) {
	p.logger.Info("Sending webhook",
		zap.String("id", req.ID),
		zap.String("url", req.URL),
		zap.String("method", req.Method),
	)

	startTime := time.Now()
	response := &WebhookResponse{
		ID:       req.ID,
		Provider: "webhook",
		SentAt:   startTime,
		Status:   "sent",
	}

	// Create HTTP request
	httpReq, err := p.createHTTPRequest(ctx, req)
	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
		response.Duration = time.Since(startTime)
		return response, err
	}

	// Send request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
		response.Duration = time.Since(startTime)
		return response, err
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		p.logger.Warn("Failed to read webhook response body", zap.Error(err))
	}

	response.StatusCode = httpResp.StatusCode
	response.ResponseBody = string(respBody)
	response.Duration = time.Since(startTime)

	// Check if request was successful
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		response.Status = "sent"
		p.logger.Info("Webhook sent successfully",
			zap.String("id", req.ID),
			zap.Int("status_code", httpResp.StatusCode),
			zap.Duration("duration", response.Duration),
		)
	} else {
		response.Status = "failed"
		response.Error = fmt.Sprintf("HTTP %d: %s", httpResp.StatusCode, httpResp.Status)
		p.logger.Warn("Webhook failed",
			zap.String("id", req.ID),
			zap.Int("status_code", httpResp.StatusCode),
			zap.String("error", response.Error),
		)
	}

	return response, nil
}

// SendBulkWebhooks sends multiple webhooks
func (p *webhookProvider) SendBulkWebhooks(ctx context.Context, requests []*WebhookRequest) ([]*WebhookResponse, error) {
	p.logger.Info("Sending bulk webhooks",
		zap.Int("count", len(requests)),
	)

	responses := make([]*WebhookResponse, len(requests))
	for i, req := range requests {
		resp, err := p.SendWebhook(ctx, req)
		if err != nil {
			// Error is already logged in SendWebhook
			resp = &WebhookResponse{
				ID:       req.ID,
				Provider: "webhook",
				SentAt:   time.Now(),
				Status:   "failed",
				Error:    err.Error(),
			}
		}
		responses[i] = resp
	}

	return responses, nil
}

// HealthCheck checks if the webhook provider is healthy
func (p *webhookProvider) HealthCheck() error {
	// Webhook provider is always healthy as it doesn't maintain persistent connections
	return nil
}

// GetProviderName returns the provider name
func (p *webhookProvider) GetProviderName() string {
	return "webhook"
}

// GetRateLimit returns the rate limit
func (p *webhookProvider) GetRateLimit() int {
	return p.config.RateLimit
}

// createHTTPRequest creates an HTTP request from webhook request
func (p *webhookProvider) createHTTPRequest(ctx context.Context, req *WebhookRequest) (*http.Request, error) {
	// Default method to POST if not specified
	method := req.Method
	if method == "" {
		method = "POST"
	}

	// Create request body
	var body io.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default content type
	contentType := req.ContentType
	if contentType == "" && req.Body != "" {
		contentType = "application/json"
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Set user agent
	httpReq.Header.Set("User-Agent", p.config.UserAgent)

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set global headers from config
	for key, value := range p.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add HMAC signature if secret is provided
	if req.Secret != "" || p.config.SigningSecret != "" {
		secret := req.Secret
		if secret == "" {
			secret = p.config.SigningSecret
		}
		signature := p.generateHMACSignature(req.Body, secret)
		httpReq.Header.Set("X-Agba-Signature", signature)
	}

	// Add timestamp header
	httpReq.Header.Set("X-Agba-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Add request ID header
	httpReq.Header.Set("X-Agba-Request-ID", req.ID)

	return httpReq, nil
}

// generateHMACSignature generates HMAC-SHA256 signature for webhook payload
func (p *webhookProvider) generateHMACSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// validateWebhookURL validates webhook URL
func (p *webhookProvider) validateWebhookURL(url string) error {
	if url == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	// In production, you might want to restrict to HTTPS only
	if p.config.VerifySSL && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must use HTTPS when SSL verification is enabled")
	}

	return nil
}

// retryWebhook retries a failed webhook with exponential backoff
func (p *webhookProvider) retryWebhook(ctx context.Context, req *WebhookRequest, attempt int) (*WebhookResponse, error) {
	// Calculate backoff delay
	delay := time.Duration(attempt) * p.config.RetryDelay
	if delay > 0 {
		p.logger.Info("Retrying webhook after delay",
			zap.String("id", req.ID),
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
		)
		
		select {
		case <-time.After(delay):
			// Continue with retry
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return p.SendWebhook(ctx, req)
}

// buildWebhookPayload builds a standard webhook payload
func (p *webhookProvider) buildWebhookPayload(eventType string, data interface{}, metadata map[string]string) map[string]interface{} {
	payload := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Unix(),
		"data":       data,
	}

	if len(metadata) > 0 {
		payload["metadata"] = metadata
	}

	return payload
}

// verifyWebhookSignature verifies HMAC signature from incoming webhook
func (p *webhookProvider) verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true // No verification if no secret
	}

	expectedSignature := p.generateHMACSignature(string(payload), secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// parseWebhookHeaders parses webhook headers from string format
func (p *webhookProvider) parseWebhookHeaders(headersStr string) map[string]string {
	headers := make(map[string]string)
	
	if headersStr == "" {
		return headers
	}

	// Simple parsing - in production you might want more robust parsing
	pairs := strings.Split(headersStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}