package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret []byte
	logger    *zap.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID       string   `json:"user_id"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	Permissions  []string `json:"permissions"`
	Organization string   `json:"organization"`
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
func (a *AuthService) ValidateAPIKey(apiKey string) (bool, error) {
	// This is a simplified implementation
	// In production, you would validate against a database or external service
	if apiKey == "" {
		return false, errors.New("API key is required")
	}

	// Add your API key validation logic here
	// For now, we'll accept any non-empty key
	return len(apiKey) > 10, nil
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