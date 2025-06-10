package handlers

import (
	"net/http"
	"strconv"

	"user-management-service/internal/models"
	"user-management-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createdBy := c.GetString("user_id")
	if createdBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req, createdBy)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUsers retrieves users with pagination
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")

	users, total, err := h.userService.GetUsers(c.Request.Context(), page, limit, search, status, role)
	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedBy := c.GetString("user_id")
	if updatedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), id, &req, updatedBy)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	deletedBy := c.GetString("user_id")
	if deletedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id, deletedBy)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.userService.ChangePassword(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to change password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService *service.UserService
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService, authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

// Login authenticates a user and returns a token
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to login", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout logs out a user
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization token required"})
		return
	}

	err := h.authService.Logout(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("Failed to logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RefreshToken refreshes an access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ForgotPassword initiates password reset process
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.authService.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process forgot password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

// ResetPassword resets a user's password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to reset password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	organizationService *service.OrganizationService
	logger              *zap.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(organizationService *service.OrganizationService, logger *zap.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		organizationService: organizationService,
		logger:              logger,
	}
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req models.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createdBy := c.GetString("user_id")
	if createdBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	organization, err := h.organizationService.CreateOrganization(c.Request.Context(), &req, createdBy)
	if err != nil {
		h.logger.Error("Failed to create organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, organization)
}

// GetOrganization retrieves an organization by ID
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
		return
	}

	id, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	organization, err := h.organizationService.GetOrganization(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get organization", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, organization)
}

// GetOrganizations retrieves organizations with pagination
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")

	organizations, total, err := h.organizationService.GetOrganizations(c.Request.Context(), page, limit, search, status)
	if err != nil {
		h.logger.Error("Failed to get organizations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve organizations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": organizations,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// UpdateOrganization updates an existing organization
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
		return
	}

	id, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req models.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedBy := c.GetString("user_id")
	if updatedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	organization, err := h.organizationService.UpdateOrganization(c.Request.Context(), id, &req, updatedBy)
	if err != nil {
		h.logger.Error("Failed to update organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	c.JSON(http.StatusOK, organization)
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
		return
	}

	id, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	deletedBy := c.GetString("user_id")
	if deletedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err = h.organizationService.DeleteOrganization(c.Request.Context(), id, deletedBy)
	if err != nil {
		h.logger.Error("Failed to delete organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// InviteUser invites a user to join an organization
func (h *OrganizationHandler) InviteUser(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
		return
	}

	id, err := uuid.Parse(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req models.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	invitedBy := c.GetString("user_id")
	if invitedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	invitation, err := h.organizationService.InviteUser(c.Request.Context(), id, &req, invitedBy)
	if err != nil {
		h.logger.Error("Failed to invite user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invite user"})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleService *service.RoleService
	logger      *zap.Logger
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService *service.RoleService, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createdBy := c.GetString("user_id")
	if createdBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	role, err := h.roleService.CreateRole(c.Request.Context(), &req, createdBy)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetRoles retrieves roles with pagination
func (h *RoleHandler) GetRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	roles, total, err := h.roleService.GetRoles(c.Request.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("Failed to get roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// AssignRole assigns a role to a user
func (h *RoleHandler) AssignRole(c *gin.Context) {
	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	assignedBy := c.GetString("user_id")
	if assignedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err := h.roleService.AssignRole(c.Request.Context(), &req, assignedBy)
	if err != nil {
		h.logger.Error("Failed to assign role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

// RevokeRole revokes a role from a user
func (h *RoleHandler) RevokeRole(c *gin.Context) {
	var req models.RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	revokedBy := c.GetString("user_id")
	if revokedBy == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err := h.roleService.RevokeRole(c.Request.Context(), &req, revokedBy)
	if err != nil {
		h.logger.Error("Failed to revoke role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role revoked successfully"})
}