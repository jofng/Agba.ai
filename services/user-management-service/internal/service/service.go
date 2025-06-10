package service

import (
    "context"
    "fmt"
    "time"

    "user-management-service/internal/repository"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

// Service handles business logic for user management operations
type Service struct {
    repository *repository.Repository
    logger     *zap.Logger
}

// NewService creates a new service instance
func NewService(repository *repository.Repository, logger *zap.Logger) *Service {
    return &Service{
        repository: repository,
        logger:     logger,
    }
}

// User represents a user
type User struct {
    ID            uuid.UUID  `json:"id"`
    Email         string     `json:"email"`
    Password      string     `json:"password,omitempty"`
    FirstName     string     `json:"first_name"`
    LastName      string     `json:"last_name"`
    Phone         string     `json:"phone"`
    Status        string     `json:"status"`
    EmailVerified bool       `json:"email_verified"`
    PhoneVerified bool       `json:"phone_verified"`
    LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, user *User) (*User, error) {
    s.logger.Info("Creating user", zap.String("email", user.Email))

    user.ID = uuid.New()
    user.Status = "active"
    user.EmailVerified = false
    user.PhoneVerified = false

    repoUser := &repository.User{
        ID:            user.ID,
        Email:         user.Email,
        Password:      user.Password,
        FirstName:     user.FirstName,
        LastName:      user.LastName,
        Phone:         user.Phone,
        Status:        user.Status,
        EmailVerified: user.EmailVerified,
        PhoneVerified: user.PhoneVerified,
    }

    if err := s.repository.CreateUser(ctx, repoUser); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    // Don't return password
    user.Password = ""
    return user, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    repoUser, err := s.repository.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }

    return &User{
        ID:            repoUser.ID,
        Email:         repoUser.Email,
        FirstName:     repoUser.FirstName,
        LastName:      repoUser.LastName,
        Phone:         repoUser.Phone,
        Status:        repoUser.Status,
        EmailVerified: repoUser.EmailVerified,
        PhoneVerified: repoUser.PhoneVerified,
        LastLoginAt:   repoUser.LastLoginAt,
        CreatedAt:     repoUser.CreatedAt,
        UpdatedAt:     repoUser.UpdatedAt,
    }, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *Service) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
    repoUser, err := s.repository.VerifyPassword(ctx, email, password)
    if err != nil {
        return nil, err
    }

    return &User{
        ID:            repoUser.ID,
        Email:         repoUser.Email,
        FirstName:     repoUser.FirstName,
        LastName:      repoUser.LastName,
        Phone:         repoUser.Phone,
        Status:        repoUser.Status,
        EmailVerified: repoUser.EmailVerified,
        PhoneVerified: repoUser.PhoneVerified,
        LastLoginAt:   repoUser.LastLoginAt,
        CreatedAt:     repoUser.CreatedAt,
        UpdatedAt:     repoUser.UpdatedAt,
    }, nil
}

// UpdateUser updates an existing user
func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, updates *User) (*User, error) {
    s.logger.Info("Updating user", zap.String("id", id.String()))

    repoUser := &repository.User{
        ID:        id,
        FirstName: updates.FirstName,
        LastName:  updates.LastName,
        Phone:     updates.Phone,
        Status:    updates.Status,
    }

    if err := s.repository.UpdateUser(ctx, repoUser); err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }

    return s.GetUser(ctx, id)
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
    s.logger.Info("Deleting user", zap.String("id", id.String()))
    return s.repository.DeleteUser(ctx, id)
}

// HealthCheck performs health checks
func (s *Service) HealthCheck(ctx context.Context) (bool, map[string]interface{}) {
    checks := make(map[string]interface{})
    
    // Check database
    if err := s.repository.HealthCheck(ctx); err != nil {
        checks["database"] = map[string]interface{}{
            "status":  "unhealthy",
            "message": err.Error(),
        }
        return false, checks
    }
    
    checks["database"] = map[string]interface{}{
        "status":  "healthy",
        "message": "Database connection successful",
    }
    
    return true, checks
}
