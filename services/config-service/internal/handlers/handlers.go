package handlers

import (
	"net/http"
	"strconv"

	"config-service/internal/models"
	"config-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConfigHandler handles configuration-related HTTP requests
type ConfigHandler struct {
	configService *service.ConfigService
	logger        *zap.Logger
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(configService *service.ConfigService, logger *zap.Logger) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
		logger:        logger,
	}
}

// CreateConfig creates a new configuration
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req models.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	config, err := h.configService.CreateConfig(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("Failed to create config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetConfig retrieves a configuration by ID
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration ID required"})
		return
	}

	id, err := uuid.Parse(configID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration ID"})
		return
	}

	config, err := h.configService.GetConfig(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get config", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetConfigs retrieves configurations with pagination
func (h *ConfigHandler) GetConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	service := c.Query("service")
	environment := c.Query("environment")

	configs, total, err := h.configService.GetConfigs(c.Request.Context(), page, limit, service, environment)
	if err != nil {
		h.logger.Error("Failed to get configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configurations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// UpdateConfig updates an existing configuration
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration ID required"})
		return
	}

	id, err := uuid.Parse(configID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration ID"})
		return
	}

	var req models.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	config, err := h.configService.UpdateConfig(c.Request.Context(), id, &req, userID)
	if err != nil {
		h.logger.Error("Failed to update config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteConfig deletes a configuration
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration ID required"})
		return
	}

	id, err := uuid.Parse(configID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration ID"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err = h.configService.DeleteConfig(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.Error("Failed to delete config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetConfigValue retrieves a specific configuration value
func (h *ConfigHandler) GetConfigValue(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")
	key := c.Param("key")

	if service == "" || environment == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service, environment, and key are required"})
		return
	}

	value, err := h.configService.GetConfigValue(c.Request.Context(), service, environment, key)
	if err != nil {
		h.logger.Error("Failed to get config value", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration value not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": value})
}

// SetConfigValue sets a specific configuration value
func (h *ConfigHandler) SetConfigValue(c *gin.Context) {
	service := c.Param("service")
	environment := c.Param("environment")
	key := c.Param("key")

	if service == "" || environment == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service, environment, and key are required"})
		return
	}

	var req struct {
		Value interface{} `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err := h.configService.SetConfigValue(c.Request.Context(), service, environment, key, req.Value, userID)
	if err != nil {
		h.logger.Error("Failed to set config value", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set configuration value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration value updated successfully"})
}

// ValidateConfig validates a configuration
func (h *ConfigHandler) ValidateConfig(c *gin.Context) {
	var req models.ValidateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := h.configService.ValidateConfig(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to validate config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate configuration"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetConfigHistory retrieves configuration history
func (h *ConfigHandler) GetConfigHistory(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration ID required"})
		return
	}

	id, err := uuid.Parse(configID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration ID"})
		return
	}

	// Get the configuration to extract service and environment
	config, err := h.configService.GetConfig(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get config", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	history, total, err := h.configService.GetConfigHistory(c.Request.Context(), config.Service, config.Environment, page, limit)
	if err != nil {
		h.logger.Error("Failed to get config history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configuration history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// TemplateHandler handles configuration template requests
type TemplateHandler struct {
	configService *service.ConfigService
	logger        *zap.Logger
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(configService *service.ConfigService, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		configService: configService,
		logger:        logger,
	}
}

// CreateTemplate creates a new configuration template
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	template, err := h.configService.CreateTemplate(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplates retrieves configuration templates
func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")

	templates, total, err := h.configService.GetTemplates(c.Request.Context(), page, limit, category)
	if err != nil {
		h.logger.Error("Failed to get templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// ApplyTemplate applies a template to create a configuration
func (h *TemplateHandler) ApplyTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	id, err := uuid.Parse(templateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req models.ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	config, err := h.configService.ApplyTemplate(c.Request.Context(), id, &req, userID)
	if err != nil {
		h.logger.Error("Failed to apply template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply template"})
		return
	}

	c.JSON(http.StatusCreated, config)
}