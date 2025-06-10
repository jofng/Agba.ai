package providers

// Package providers implements notification delivery providers for email, SMS, push notifications, and webhooks.
// Each provider implements a common interface and supports multiple backend services (e.g., SMTP, SendGrid for email).

// This file serves as the main entry point for the providers package.
// Individual provider implementations are in separate files:
// - email.go: Email provider implementations (SMTP, SendGrid, SES, Mailgun)
// - sms.go: SMS provider implementations (Twilio, AWS SNS, Nexmo)
// - push.go: Push notification providers (FCM, APNS)
// - webhook.go: Webhook provider implementation
// - interfaces.go: Common interfaces and data structures

// All provider factory functions are implemented in their respective files.