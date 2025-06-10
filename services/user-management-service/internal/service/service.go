package service

import (
	"context"
	"fmt"
	"time"

	"user-management-service/internal/auth"
	"user-management-service/internal/models"
	"user-management-service/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service interfaces
type UserServiceInterface interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, organizationID *uuid.UUID, limit, offset int) ([]*models.User, int, error)
	ActivateUser(ctx context.Context, id uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	VerifyEmail(ctx context.Context, token string) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID, assignedBy uuid.UUID) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type OrganizationServiceInterface interface {
	CreateOrganization(ctx context.Context, org *models.Organization) (*models.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	UpdateOrganization(ctx context.Context, org *models.Organization) (*models.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	ListOrganizations(ctx context.Context, limit, offset int) ([]*models.Organization, int, error)
	GetOrganizationUsers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.User, int, error)
	AddUser(ctx context.Context, orgID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, orgID, userID uuid.UUID) error
	InviteUser(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error)
	GetInvitations(ctx context.Context, orgID uuid.UUID) ([]*models.Invitation, error)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error
	DeclineInvitation(ctx context.Context, token string) error
}

type RoleServiceInterface interface {
	CreateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListRoles(ctx context.Context, organizationID *uuid.UUID, limit, offset int) ([]*models.Role, int, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
	AddPermission(ctx context.Context, roleID uuid.UUID, permission string) error
	RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error
	GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*models.User, int, error)
}

type SessionServiceInterface interface {
	CreateSession(ctx context.Context, session *models.Session) (*models.Session, error)
	GetSession(ctx context.Context, id string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) (*models.Session, error)
	DeleteSession(ctx context.Context, id string) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	RevokeSession(ctx context.Context, id string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

type AuditServiceInterface interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	GetAuditLog(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
	GetAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.AuditLog, int, error)
	GetUserAuditLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error)
	GetOrganizationAuditLogs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error)
}

// UserService implementation
type UserService struct {
	userRepo         repository.UserRepositoryInterface
	organizationRepo repository.OrganizationRepositoryInterface
	roleRepo         repository.RoleRepositoryInterface
	authManager      *auth.Manager
	natsConn         repository.NATSInterface
	logger           *zap.Logger
}

func NewUserService(
	userRepo repository.UserRepositoryInterface,
	organizationRepo repository.OrganizationRepositoryInterface,
	roleRepo repository.RoleRepositoryInterface,
	authManager *auth.Manager,
	natsConn repository.NATSInterface,
	logger *zap.Logger,
) UserServiceInterface {
	return &UserService{
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		roleRepo:         roleRepo,
		authManager:      authManager,
		natsConn:         natsConn,
		logger:           logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	s.logger.Info("Creating user", zap.String("email", user.Email))

	user.ID = uuid.New()
	user.Status = models.UserStatusActive
	user.EmailVerified = false
	user.PhoneVerified = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUser(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	user.UpdatedAt = time.Now()
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.DeleteUser(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, organizationID *uuid.UUID, limit, offset int) ([]*models.User, int, error) {
	return s.userRepo.ListUsers(ctx, organizationID, limit, offset)
}

func (s *UserService) ActivateUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetUser(ctx, id)
	if err != nil {
		return err
	}
	user.Status = models.UserStatusActive
	user.UpdatedAt = time.Now()
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *UserService) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetUser(ctx, id)
	if err != nil {
		return err
	}
	user.Status = models.UserStatusInactive
	user.UpdatedAt = time.Now()
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// TODO: Implement password change logic
	s.logger.Info("Changing password", zap.String("user_id", userID.String()))
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// TODO: Implement password reset logic
	s.logger.Info("Resetting password", zap.String("token", token))
	return nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement email verification logic
	s.logger.Info("Verifying email", zap.String("token", token))
	return nil
}

func (s *UserService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	return s.userRepo.GetUserRoles(ctx, userID)
}

func (s *UserService) AssignRole(ctx context.Context, userID, roleID uuid.UUID, assignedBy uuid.UUID) error {
	return s.userRepo.AssignRole(ctx, userID, roleID, assignedBy)
}

func (s *UserService) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.userRepo.RemoveRole(ctx, userID, roleID)
}

func (s *UserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.userRepo.GetUserPermissions(ctx, userID)
}

// OrganizationService implementation
type OrganizationService struct {
	organizationRepo repository.OrganizationRepositoryInterface
	userRepo         repository.UserRepositoryInterface
	roleRepo         repository.RoleRepositoryInterface
	natsConn         repository.NATSInterface
	logger           *zap.Logger
}

func NewOrganizationService(
	organizationRepo repository.OrganizationRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	roleRepo repository.RoleRepositoryInterface,
	natsConn repository.NATSInterface,
	logger *zap.Logger,
) OrganizationServiceInterface {
	return &OrganizationService{
		organizationRepo: organizationRepo,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		natsConn:         natsConn,
		logger:           logger,
	}
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, org *models.Organization) (*models.Organization, error) {
	org.ID = uuid.New()
	org.Status = models.OrganizationStatusActive
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()

	if err := s.organizationRepo.CreateOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	return s.organizationRepo.GetOrganization(ctx, id)
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, org *models.Organization) (*models.Organization, error) {
	org.UpdatedAt = time.Now()
	if err := s.organizationRepo.UpdateOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}
	return org, nil
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return s.organizationRepo.DeleteOrganization(ctx, id)
}

func (s *OrganizationService) ListOrganizations(ctx context.Context, limit, offset int) ([]*models.Organization, int, error) {
	return s.organizationRepo.ListOrganizations(ctx, limit, offset)
}

func (s *OrganizationService) GetOrganizationUsers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.User, int, error) {
	return s.organizationRepo.GetOrganizationUsers(ctx, orgID, limit, offset)
}

func (s *OrganizationService) AddUser(ctx context.Context, orgID, userID uuid.UUID) error {
	return s.organizationRepo.AddUser(ctx, orgID, userID)
}

func (s *OrganizationService) RemoveUser(ctx context.Context, orgID, userID uuid.UUID) error {
	return s.organizationRepo.RemoveUser(ctx, orgID, userID)
}

func (s *OrganizationService) InviteUser(ctx context.Context, invitation *models.Invitation) (*models.Invitation, error) {
	invitation.ID = uuid.New()
	invitation.Status = models.InvitationStatusPending
	invitation.CreatedAt = time.Now()
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days

	if err := s.organizationRepo.CreateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return invitation, nil
}

func (s *OrganizationService) GetInvitations(ctx context.Context, orgID uuid.UUID) ([]*models.Invitation, error) {
	return s.organizationRepo.GetInvitations(ctx, orgID)
}

func (s *OrganizationService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	invitation, err := s.organizationRepo.GetInvitation(ctx, token)
	if err != nil {
		return err
	}

	invitation.Status = models.InvitationStatusAccepted
	now := time.Now()
	invitation.AcceptedAt = &now

	return s.organizationRepo.UpdateInvitation(ctx, invitation)
}

func (s *OrganizationService) DeclineInvitation(ctx context.Context, token string) error {
	invitation, err := s.organizationRepo.GetInvitation(ctx, token)
	if err != nil {
		return err
	}

	invitation.Status = models.InvitationStatusDeclined
	now := time.Now()
	invitation.DeclinedAt = &now

	return s.organizationRepo.UpdateInvitation(ctx, invitation)
}

// RoleService implementation
type RoleService struct {
	roleRepo repository.RoleRepositoryInterface
	userRepo repository.UserRepositoryInterface
	logger   *zap.Logger
}

func NewRoleService(
	roleRepo repository.RoleRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	logger *zap.Logger,
) RoleServiceInterface {
	return &RoleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *RoleService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return s.roleRepo.GetRole(ctx, id)
}

func (s *RoleService) UpdateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	role.UpdatedAt = time.Now()
	if err := s.roleRepo.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return role, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.roleRepo.DeleteRole(ctx, id)
}

func (s *RoleService) ListRoles(ctx context.Context, organizationID *uuid.UUID, limit, offset int) ([]*models.Role, int, error) {
	return s.roleRepo.ListRoles(ctx, organizationID, limit, offset)
}

func (s *RoleService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

func (s *RoleService) AddPermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return s.roleRepo.AddPermission(ctx, roleID, permission)
}

func (s *RoleService) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return s.roleRepo.RemovePermission(ctx, roleID, permission)
}

func (s *RoleService) GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*models.User, int, error) {
	return s.roleRepo.GetRoleUsers(ctx, roleID, limit, offset)
}

// SessionService implementation
type SessionService struct {
	sessionRepo repository.SessionRepositoryInterface
	userRepo    repository.UserRepositoryInterface
	authManager *auth.Manager
	logger      *zap.Logger
}

func NewSessionService(
	sessionRepo repository.SessionRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	authManager *auth.Manager,
	logger *zap.Logger,
) SessionServiceInterface {
	return &SessionService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		authManager: authManager,
		logger:      logger,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, session *models.Session) (*models.Session, error) {
	session.ID = uuid.New()
	session.Status = models.SessionStatusActive
	session.CreatedAt = time.Now()
	session.LastAccessedAt = time.Now()
	session.ExpiresAt = time.Now().Add(24 * time.Hour) // 24 hours

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *SessionService) GetSession(ctx context.Context, id string) (*models.Session, error) {
	return s.sessionRepo.GetSession(ctx, id)
}

func (s *SessionService) UpdateSession(ctx context.Context, session *models.Session) (*models.Session, error) {
	session.LastAccessedAt = time.Now()
	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	return session, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, id string) error {
	return s.sessionRepo.DeleteSession(ctx, id)
}

func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	return s.sessionRepo.GetUserSessions(ctx, userID)
}

func (s *SessionService) RevokeSession(ctx context.Context, id string) error {
	session, err := s.sessionRepo.GetSession(ctx, id)
	if err != nil {
		return err
	}

	session.Status = models.SessionStatusRevoked
	now := time.Now()
	session.RevokedAt = &now

	return s.sessionRepo.UpdateSession(ctx, session)
}

func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteUserSessions(ctx, userID)
}

func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.CleanupExpiredSessions(ctx)
}

// AuditService implementation
type AuditService struct {
	auditRepo repository.AuditRepositoryInterface
	logger    *zap.Logger
}

func NewAuditService(
	auditRepo repository.AuditRepositoryInterface,
	logger *zap.Logger,
) AuditServiceInterface {
	return &AuditService{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

func (s *AuditService) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	log.ID = uuid.New()
	log.Timestamp = time.Now()
	return s.auditRepo.CreateAuditLog(ctx, log)
}

func (s *AuditService) GetAuditLog(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	return s.auditRepo.GetAuditLog(ctx, id)
}

func (s *AuditService) GetAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.AuditLog, int, error) {
	return s.auditRepo.GetAuditLogs(ctx, filters, limit, offset)
}

func (s *AuditService) GetUserAuditLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error) {
	return s.auditRepo.GetUserAuditLogs(ctx, userID, limit, offset)
}

func (s *AuditService) GetOrganizationAuditLogs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error) {
	return s.auditRepo.GetOrganizationAuditLogs(ctx, orgID, limit, offset)
}
