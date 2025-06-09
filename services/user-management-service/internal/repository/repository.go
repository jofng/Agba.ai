package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/agba-ai/user-management-service/internal/models"
    "github.com/google/uuid"
    "go.uber.org/zap"
    "golang.org/x/crypto/bcrypt"
)

// Repository handles data operations for user management
type Repository struct {
    db     *sql.DB
    logger *zap.Logger
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
    return &Repository{
        db:     db,
        logger: logger,
    }
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    query := `
        INSERT INTO users (id, email, password_hash, first_name, last_name, phone, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

    now := time.Now()
    _, err = r.db.ExecContext(ctx, query,
        user.ID, user.Email, string(hashedPassword), user.FirstName,
        user.LastName, user.Phone, user.Status, now, now)

    if err != nil {
        r.logger.Error("Failed to create user", zap.Error(err))
        return fmt.Errorf("failed to create user: %w", err)
    }

    return nil
}

// GetUser retrieves a user by ID
func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
    query := `
        SELECT id, email, first_name, last_name, phone, status, email_verified, phone_verified, last_login_at, created_at, updated_at
        FROM users WHERE id = $1`

    var user models.User
    var lastLoginAt sql.NullTime

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.FirstName, &user.LastName,
        &user.Phone, &user.Status, &user.EmailVerified, &user.PhoneVerified,
        &lastLoginAt, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found: %s", id.String())
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if lastLoginAt.Valid {
        user.LastLoginAt = &lastLoginAt.Time
    }

    return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    query := `
        SELECT id, email, password_hash, first_name, last_name, phone, status, email_verified, phone_verified, last_login_at, created_at, updated_at
        FROM users WHERE email = $1`

    var user models.User
    var lastLoginAt sql.NullTime

    err := r.db.QueryRowContext(ctx, query, email).Scan(
        &user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
        &user.Phone, &user.Status, &user.EmailVerified, &user.PhoneVerified,
        &lastLoginAt, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found: %s", email)
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if lastLoginAt.Valid {
        user.LastLoginAt = &lastLoginAt.Time
    }

    return &user, nil
}

// UpdateUser updates an existing user
func (r *Repository) UpdateUser(ctx context.Context, user *models.User) error {
    query := `
        UPDATE users 
        SET first_name = $2, last_name = $3, phone = $4, status = $5, updated_at = $6
        WHERE id = $1`

    _, err := r.db.ExecContext(ctx, query,
        user.ID, user.FirstName, user.LastName, user.Phone, user.Status, time.Now())

    if err != nil {
        r.logger.Error("Failed to update user", zap.Error(err))
        return fmt.Errorf("failed to update user: %w", err)
    }

    return nil
}

// DeleteUser deletes a user
func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
    query := "DELETE FROM users WHERE id = $1"
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        r.logger.Error("Failed to delete user", zap.Error(err))
        return fmt.Errorf("failed to delete user: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("user not found: %s", id.String())
    }

    return nil
}

// VerifyPassword verifies a user's password
func (r *Repository) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
    user, err := r.GetUserByEmail(ctx, email)
    if err != nil {
        return nil, err
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        return nil, fmt.Errorf("invalid password")
    }

    return user, nil
}

// HealthCheck checks repository health
func (r *Repository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
