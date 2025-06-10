package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

// Manager handles authentication operations
type Manager struct {
	config *Config
	logger *zap.Logger
}

// Config holds authentication configuration
type Config struct {
	JWTSecret           string        `mapstructure:"jwt_secret"`
	JWTExpiration       time.Duration `mapstructure:"jwt_expiration"`
	RefreshExpiration   time.Duration `mapstructure:"refresh_expiration"`
	PasswordMinLength   int           `mapstructure:"password_min_length"`
	PasswordRequireUpper bool         `mapstructure:"password_require_upper"`
	PasswordRequireLower bool         `mapstructure:"password_require_lower"`
	PasswordRequireDigit bool         `mapstructure:"password_require_digit"`
	PasswordRequireSpecial bool       `mapstructure:"password_require_special"`
	MaxLoginAttempts    int           `mapstructure:"max_login_attempts"`
	LockoutDuration     time.Duration `mapstructure:"lockout_duration"`
}

// Claims represents JWT claims
type Claims struct {
	UserID         string   `json:"user_id"`
	OrganizationID string   `json:"organization_id"`
	Email          string   `json:"email"`
	Roles          []string `json:"roles"`
	Permissions    []string `json:"permissions"`
	SessionID      string   `json:"session_id"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewManager creates a new authentication manager
func NewManager(config *Config, logger *zap.Logger) *Manager {
	return &Manager{
		config: config,
		logger: logger,
	}
}

// HashPassword hashes a password using bcrypt
func (m *Manager) HashPassword(password string) (string, error) {
	if err := m.ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		m.logger.Error("Failed to hash password", zap.Error(err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (m *Manager) VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid password")
		}
		m.logger.Error("Failed to verify password", zap.Error(err))
		return fmt.Errorf("failed to verify password: %w", err)
	}
	return nil
}

// ValidatePassword validates password strength
func (m *Manager) ValidatePassword(password string) error {
	if len(password) < m.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", m.config.PasswordMinLength)
	}

	if m.config.PasswordRequireUpper {
		hasUpper := false
		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return errors.New("password must contain at least one uppercase letter")
		}
	}

	if m.config.PasswordRequireLower {
		hasLower := false
		for _, char := range password {
			if char >= 'a' && char <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return errors.New("password must contain at least one lowercase letter")
		}
	}

	if m.config.PasswordRequireDigit {
		hasDigit := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasDigit = true
				break
			}
		}
		if !hasDigit {
			return errors.New("password must contain at least one digit")
		}
	}

	if m.config.PasswordRequireSpecial {
		hasSpecial := false
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		for _, char := range password {
			for _, special := range specialChars {
				if char == special {
					hasSpecial = true
					break
				}
			}
			if hasSpecial {
				break
			}
		}
		if !hasSpecial {
			return errors.New("password must contain at least one special character")
		}
	}

	return nil
}

// GenerateTokenPair generates access and refresh tokens
func (m *Manager) GenerateTokenPair(userID, organizationID, email string, roles, permissions []string, sessionID string) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.JWTExpiration)

	// Create access token claims
	claims := &Claims{
		UserID:         userID,
		OrganizationID: organizationID,
		Email:          email,
		Roles:          roles,
		Permissions:    permissions,
		SessionID:      sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "agba-user-management",
			Subject:   userID,
		},
	}

	// Generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(m.config.JWTSecret))
	if err != nil {
		m.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := m.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates and parses a JWT token
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		m.logger.Debug("Token validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// RefreshToken generates a new access token using a refresh token
func (m *Manager) RefreshToken(refreshToken string, userID, organizationID, email string, roles, permissions []string, sessionID string) (*TokenPair, error) {
	// In a real implementation, you would validate the refresh token against stored tokens
	// For now, we'll just generate a new token pair
	return m.GenerateTokenPair(userID, organizationID, email, roles, permissions, sessionID)
}

// generateRefreshToken generates a secure random refresh token
func (m *Manager) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		m.logger.Error("Failed to generate refresh token", zap.Error(err))
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateResetToken generates a password reset token
func (m *Manager) GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		m.logger.Error("Failed to generate reset token", zap.Error(err))
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateVerificationToken generates an email verification token
func (m *Manager) GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		m.logger.Error("Failed to generate verification token", zap.Error(err))
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HasPermission checks if the user has a specific permission
func (m *Manager) HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission || permission == "*" {
			return true
		}
	}
	return false
}

// HasRole checks if the user has a specific role
func (m *Manager) HasRole(userRoles []string, requiredRole string) bool {
	for _, role := range userRoles {
		if role == requiredRole || role == "admin" {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (m *Manager) HasAnyRole(userRoles []string, requiredRoles []string) bool {
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole || userRole == "admin" {
				return true
			}
		}
	}
	return false
}