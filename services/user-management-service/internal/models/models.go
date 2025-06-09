package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
	UserStatusDeleted   UserStatus = "deleted"
)

// OrganizationStatus represents the status of an organization
type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusInactive  OrganizationStatus = "inactive"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusDeleted   OrganizationStatus = "deleted"
)

// RoleType represents the type of role
type RoleType string

const (
	RoleTypeSystem       RoleType = "system"
	RoleTypeOrganization RoleType = "organization"
	RoleTypeCustom       RoleType = "custom"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	Username       string                 `json:"username" db:"username"`
	Email          string                 `json:"email" db:"email"`
	EmailVerified  bool                   `json:"email_verified" db:"email_verified"`
	PasswordHash   string                 `json:"-" db:"password_hash"`
	FirstName      string                 `json:"first_name" db:"first_name"`
	LastName       string                 `json:"last_name" db:"last_name"`
	DisplayName    string                 `json:"display_name" db:"display_name"`
	Avatar         string                 `json:"avatar,omitempty" db:"avatar"`
	Phone          string                 `json:"phone,omitempty" db:"phone"`
	PhoneVerified  bool                   `json:"phone_verified" db:"phone_verified"`
	Status         UserStatus             `json:"status" db:"status"`
	Language       string                 `json:"language" db:"language"`
	Timezone       string                 `json:"timezone" db:"timezone"`
	
	// Organization relationship
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty" db:"organization_id"`
	
	// Authentication settings
	TwoFactorEnabled    bool              `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret     string            `json:"-" db:"two_factor_secret"`
	BackupCodes         []string          `json:"-" db:"backup_codes"`
	
	// Account security
	LastLoginAt         *time.Time        `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP         string            `json:"last_login_ip,omitempty" db:"last_login_ip"`
	FailedLoginAttempts int               `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time        `json:"locked_until,omitempty" db:"locked_until"`
	PasswordChangedAt   *time.Time        `json:"password_changed_at,omitempty" db:"password_changed_at"`
	
	// Verification tokens
	EmailVerificationToken    string     `json:"-" db:"email_verification_token"`
	EmailVerificationExpiry   *time.Time `json:"-" db:"email_verification_expiry"`
	PasswordResetToken        string     `json:"-" db:"password_reset_token"`
	PasswordResetExpiry       *time.Time `json:"-" db:"password_reset_expiry"`
	
	// User preferences
	Preferences map[string]interface{}    `json:"preferences,omitempty" db:"preferences"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time                `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Organization represents an organization in the system
type Organization struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	DisplayName string                    `json:"display_name" db:"display_name"`
	Description string                    `json:"description,omitempty" db:"description"`
	Website     string                    `json:"website,omitempty" db:"website"`
	Logo        string                    `json:"logo,omitempty" db:"logo"`
	Status      OrganizationStatus        `json:"status" db:"status"`
	
	// Contact information
	Email       string                    `json:"email,omitempty" db:"email"`
	Phone       string                    `json:"phone,omitempty" db:"phone"`
	Address     Address                   `json:"address,omitempty" db:"address"`
	
	// Organization settings
	Settings    OrganizationSettings      `json:"settings" db:"settings"`
	
	// Subscription and billing
	SubscriptionPlan   string             `json:"subscription_plan,omitempty" db:"subscription_plan"`
	SubscriptionStatus string             `json:"subscription_status,omitempty" db:"subscription_status"`
	BillingEmail       string             `json:"billing_email,omitempty" db:"billing_email"`
	
	// Limits and quotas
	UserLimit          int                `json:"user_limit" db:"user_limit"`
	StorageLimit       int64              `json:"storage_limit" db:"storage_limit"`
	APICallLimit       int                `json:"api_call_limit" db:"api_call_limit"`
	
	// Owner information
	OwnerID     uuid.UUID                 `json:"owner_id" db:"owner_id"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time                `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Address represents a physical address
type Address struct {
	Street1    string `json:"street1,omitempty"`
	Street2    string `json:"street2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// OrganizationSettings represents organization-specific settings
type OrganizationSettings struct {
	AllowSelfRegistration   bool     `json:"allow_self_registration"`
	RequireEmailVerification bool    `json:"require_email_verification"`
	AllowedDomains          []string `json:"allowed_domains,omitempty"`
	DefaultRole             string   `json:"default_role"`
	SessionTimeout          int      `json:"session_timeout"` // in minutes
	PasswordPolicy          PasswordPolicy `json:"password_policy"`
	TwoFactorRequired       bool     `json:"two_factor_required"`
	IPWhitelist             []string `json:"ip_whitelist,omitempty"`
}

// PasswordPolicy represents password policy settings
type PasswordPolicy struct {
	MinLength      int  `json:"min_length"`
	MaxLength      int  `json:"max_length"`
	RequireUpper   bool `json:"require_upper"`
	RequireLower   bool `json:"require_lower"`
	RequireDigit   bool `json:"require_digit"`
	RequireSpecial bool `json:"require_special"`
	MaxAge         int  `json:"max_age"` // in days
	HistoryCount   int  `json:"history_count"`
}

// Role represents a role in the system
type Role struct {
	ID             uuid.UUID                 `json:"id" db:"id"`
	Name           string                    `json:"name" db:"name"`
	DisplayName    string                    `json:"display_name" db:"display_name"`
	Description    string                    `json:"description,omitempty" db:"description"`
	Type           RoleType                  `json:"type" db:"type"`
	OrganizationID *uuid.UUID                `json:"organization_id,omitempty" db:"organization_id"`
	
	// Role configuration
	Permissions    []string                  `json:"permissions" db:"permissions"`
	IsDefault      bool                      `json:"is_default" db:"is_default"`
	IsSystem       bool                      `json:"is_system" db:"is_system"`
	
	// Metadata
	Metadata       map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at" db:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	DisplayName string                    `json:"display_name" db:"display_name"`
	Description string                    `json:"description,omitempty" db:"description"`
	Resource    string                    `json:"resource" db:"resource"`
	Action      string                    `json:"action" db:"action"`
	Scope       string                    `json:"scope" db:"scope"`
	
	// Metadata
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// UserRole represents the relationship between a user and a role
type UserRole struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	UserID         uuid.UUID     `json:"user_id" db:"user_id"`
	RoleID         uuid.UUID     `json:"role_id" db:"role_id"`
	OrganizationID *uuid.UUID    `json:"organization_id,omitempty" db:"organization_id"`
	AssignedBy     uuid.UUID     `json:"assigned_by" db:"assigned_by"`
	AssignedAt     time.Time     `json:"assigned_at" db:"assigned_at"`
	ExpiresAt      *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
}

// Session represents a user session
type Session struct {
	ID             uuid.UUID                 `json:"id" db:"id"`
	UserID         uuid.UUID                 `json:"user_id" db:"user_id"`
	Token          string                    `json:"-" db:"token"`
	RefreshToken   string                    `json:"-" db:"refresh_token"`
	Status         SessionStatus             `json:"status" db:"status"`
	
	// Session metadata
	IPAddress      string                    `json:"ip_address" db:"ip_address"`
	UserAgent      string                    `json:"user_agent" db:"user_agent"`
	DeviceInfo     DeviceInfo                `json:"device_info,omitempty" db:"device_info"`
	Location       Location                  `json:"location,omitempty" db:"location"`
	
	// Session timing
	CreatedAt      time.Time                 `json:"created_at" db:"created_at"`
	LastAccessedAt time.Time                 `json:"last_accessed_at" db:"last_accessed_at"`
	ExpiresAt      time.Time                 `json:"expires_at" db:"expires_at"`
	RevokedAt      *time.Time                `json:"revoked_at,omitempty" db:"revoked_at"`
}

// DeviceInfo represents device information for a session
type DeviceInfo struct {
	Type     string `json:"type,omitempty"`     // mobile, desktop, tablet
	OS       string `json:"os,omitempty"`       // iOS, Android, Windows, macOS, Linux
	Browser  string `json:"browser,omitempty"`  // Chrome, Firefox, Safari, Edge
	Version  string `json:"version,omitempty"`  // Browser/OS version
}

// Location represents geographical location information
type Location struct {
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// Invitation represents an invitation to join an organization
type Invitation struct {
	ID             uuid.UUID                 `json:"id" db:"id"`
	Email          string                    `json:"email" db:"email"`
	OrganizationID uuid.UUID                 `json:"organization_id" db:"organization_id"`
	RoleID         uuid.UUID                 `json:"role_id" db:"role_id"`
	InvitedBy      uuid.UUID                 `json:"invited_by" db:"invited_by"`
	Status         InvitationStatus          `json:"status" db:"status"`
	Token          string                    `json:"-" db:"token"`
	Message        string                    `json:"message,omitempty" db:"message"`
	
	// Invitation timing
	CreatedAt      time.Time                 `json:"created_at" db:"created_at"`
	ExpiresAt      time.Time                 `json:"expires_at" db:"expires_at"`
	AcceptedAt     *time.Time                `json:"accepted_at,omitempty" db:"accepted_at"`
	DeclinedAt     *time.Time                `json:"declined_at,omitempty" db:"declined_at"`
	
	// Metadata
	Metadata       map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             uuid.UUID                 `json:"id" db:"id"`
	UserID         *uuid.UUID                `json:"user_id,omitempty" db:"user_id"`
	OrganizationID *uuid.UUID                `json:"organization_id,omitempty" db:"organization_id"`
	Action         string                    `json:"action" db:"action"`
	Resource       string                    `json:"resource" db:"resource"`
	ResourceID     *uuid.UUID                `json:"resource_id,omitempty" db:"resource_id"`
	
	// Request information
	IPAddress      string                    `json:"ip_address" db:"ip_address"`
	UserAgent      string                    `json:"user_agent" db:"user_agent"`
	
	// Change details
	OldValues      map[string]interface{}    `json:"old_values,omitempty" db:"old_values"`
	NewValues      map[string]interface{}    `json:"new_values,omitempty" db:"new_values"`
	
	// Result
	Success        bool                      `json:"success" db:"success"`
	ErrorMessage   string                    `json:"error_message,omitempty" db:"error_message"`
	
	// Metadata
	Metadata       map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	Timestamp      time.Time                 `json:"timestamp" db:"timestamp"`
}

// UserProfile represents a user's profile information
type UserProfile struct {
	UserID      uuid.UUID                 `json:"user_id"`
	Bio         string                    `json:"bio,omitempty"`
	Website     string                    `json:"website,omitempty"`
	Location    string                    `json:"location,omitempty"`
	Company     string                    `json:"company,omitempty"`
	JobTitle    string                    `json:"job_title,omitempty"`
	SocialLinks map[string]string         `json:"social_links,omitempty"`
	Skills      []string                  `json:"skills,omitempty"`
	Interests   []string                  `json:"interests,omitempty"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// UserActivity represents user activity information
type UserActivity struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	UserID      uuid.UUID                 `json:"user_id" db:"user_id"`
	Action      string                    `json:"action" db:"action"`
	Description string                    `json:"description" db:"description"`
	IPAddress   string                    `json:"ip_address" db:"ip_address"`
	UserAgent   string                    `json:"user_agent" db:"user_agent"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	Timestamp   time.Time                 `json:"timestamp" db:"timestamp"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
	TwoFactorCode string `json:"two_factor_code,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	User         *User     `json:"user"`
	Permissions  []string  `json:"permissions"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	OrganizationName string `json:"organization_name,omitempty"`
	InvitationToken  string `json:"invitation_token,omitempty"`
	AcceptTerms     bool   `json:"accept_terms" binding:"required"`
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation request
type PasswordResetConfirmRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// EmailVerificationRequest represents an email verification request
type EmailVerificationRequest struct {
	Token string `json:"token" binding:"required"`
}

// InviteUserRequest represents a user invitation request
type InviteUserRequest struct {
	Email   string    `json:"email" binding:"required,email"`
	RoleID  uuid.UUID `json:"role_id" binding:"required"`
	Message string    `json:"message,omitempty"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Language    *string `json:"language,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	Status      *UserStatus `json:"status,omitempty"`
}

// UpdateOrganizationRequest represents an organization update request
type UpdateOrganizationRequest struct {
	Name        *string                   `json:"name,omitempty"`
	DisplayName *string                   `json:"display_name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Website     *string                   `json:"website,omitempty"`
	Email       *string                   `json:"email,omitempty"`
	Phone       *string                   `json:"phone,omitempty"`
	Address     *Address                  `json:"address,omitempty"`
	Settings    *OrganizationSettings     `json:"settings,omitempty"`
}

// CreateRoleRequest represents a role creation request
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleRequest represents a role update request
type UpdateRoleRequest struct {
	DisplayName *string  `json:"display_name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// SystemStats represents system statistics
type SystemStats struct {
	TotalUsers         int64     `json:"total_users"`
	ActiveUsers        int64     `json:"active_users"`
	TotalOrganizations int64     `json:"total_organizations"`
	ActiveSessions     int64     `json:"active_sessions"`
	NewUsersToday      int64     `json:"new_users_today"`
	NewUsersThisWeek   int64     `json:"new_users_this_week"`
	NewUsersThisMonth  int64     `json:"new_users_this_month"`
	LastUpdated        time.Time `json:"last_updated"`
}