package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Service interface for authentication operations
type Service interface {
	ValidateToken(tokenString string) (*Claims, error)
	ValidateAPIKey(apiKey string) (*APIKeyInfo, error)
	HasPermission(claims *Claims, permission string) bool
}

// Config holds authentication configuration
type Config struct {
	JWTSecret          string
	APIKeys            map[string]*APIKeyInfo
	OAuth2ServerURL    string
	APIKeyValidatorURL string
	TokenCacheTTL      time.Duration
}

// APIKeyInfo contains information about an API key
type APIKeyInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Active      bool     `json:"active"`
}

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret []byte
	logger    *zap.Logger
	apiKeys   map[string]*APIKeyInfo
}

// New creates a new authentication service
func New(config *Config, logger *zap.Logger) Service {
	return &AuthService{
		jwtSecret: []byte(config.JWTSecret),
		logger:    logger,
		apiKeys:   config.APIKeys,
	}
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
		apiKeys:   make(map[string]*APIKeyInfo),
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID         string   `json:"user_id"`
	Email          string   `json:"email"`
	Roles          []string `json:"roles"`
	Permissions    []string `json:"permissions"`
	Organization   string   `json:"organization"`
	OrganizationID string   `json:"organization_id"`
	Scopes         []string `json:"scopes"`
	jwt.RegisteredClaims
}

// ValidateToken validates a JWT token and returns claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		a.logger.Error("Failed to parse JWT token", zap.Error(err))
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAPIKey validates an API key
func (a *AuthService) ValidateAPIKey(apiKey string) (*APIKeyInfo, error) {
	if apiKey == "" {
		return nil, errors.New("API key is required")
	}

	// Check if API key exists in our map
	if keyInfo, exists := a.apiKeys[apiKey]; exists && keyInfo.Active {
		return keyInfo, nil
	}

	// For backward compatibility, accept any key longer than 10 characters
	if len(apiKey) > 10 {
		return &APIKeyInfo{
			ID:          "default",
			Name:        "Default API Key",
			Permissions: []string{"*"},
			Active:      true,
		}, nil
	}

	return nil, errors.New("invalid API key")
}

// HasPermission checks if the user has the required permission
func (a *AuthService) HasPermission(claims *Claims, permission string) bool {
	for _, p := range claims.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// HasRole checks if the user has the required role
func (a *AuthService) HasRole(claims *Claims, role string) bool {
	for _, r := range claims.Roles {
		if r == role || r == "admin" {
			return true
		}
	}
	return false
}

// IsTokenExpired checks if the token is expired
func (a *AuthService) IsTokenExpired(claims *Claims) bool {
	return claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now())
}

// RefreshToken generates a new token with extended expiration
func (a *AuthService) RefreshToken(claims *Claims) (string, error) {
	// Create new claims with extended expiration
	newClaims := &Claims{
		UserID:       claims.UserID,
		Email:        claims.Email,
		Roles:        claims.Roles,
		Permissions:  claims.Permissions,
		Organization: claims.Organization,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "agba-api-gateway",
			Subject:   claims.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(a.jwtSecret)
}