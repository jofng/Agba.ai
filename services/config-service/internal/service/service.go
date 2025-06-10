package service

import (
	"context"
	"encoding/json"
	"fmt"

	"config-service/internal/models"
	"config-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConfigService handles business logic for configuration management
type ConfigService struct {
	dbRepo    repository.DatabaseRepository
	redisRepo repository.RedisRepository
	natsRepo  repository.NATSRepository
	logger    *zap.Logger
}

// NewConfigService creates a new configuration service
func NewConfigService(
	dbRepo repository.DatabaseRepository,
	redisRepo repository.RedisRepository,
	natsRepo repository.NATSRepository,
	logger *zap.Logger,
) *ConfigService {
	return &ConfigService{
		dbRepo:    dbRepo,
		redisRepo: redisRepo,
		natsRepo:  natsRepo,
		logger:    logger,
	}
}

// CreateConfig creates a new configuration
func (s *ConfigService) CreateConfig(ctx context.Context, req *models.CreateConfigRequest, userID string) (*models.Configuration, error) {
	// Validate the configuration
	if err := s.validateConfigData(req.Data); err != nil {
		return nil, fmt.Errorf("invalid configuration data: %w", err)
	}

	config := &models.Configuration{
		ID:          uuid.New(),
		Service:     req.Service,
		Environment: req.Environment,
		Type:        req.Type,
		EntityID:    req.EntityID,
		Name:        req.Name,
		Description: req.Description,
		Version:     1, // Start with version 1
		Status:      models.ConfigStatusDraft,
		Data:        req.Data,
		Schema:      req.Schema,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		CreatedBy:   uuid.MustParse(userID),
		UpdatedBy:   uuid.MustParse(userID),
	}

	// Save to database
	if err := s.dbRepo.CreateConfiguration(ctx, config); err != nil {
		s.logger.Error("Failed to create config in database", zap.Error(err))
		return nil, err
	}

	// Cache the configuration
	if err := s.cacheConfig(ctx, config); err != nil {
		s.logger.Warn("Failed to cache config", zap.Error(err))
	}

	// Publish configuration change event
	if err := s.publishConfigEvent(ctx, "config.created", config); err != nil {
		s.logger.Warn("Failed to publish config event", zap.Error(err))
	}

	s.logger.Info("Configuration created",
		zap.String("config_id", config.ID.String()),
		zap.String("service", config.Service),
		zap.String("environment", config.Environment),
	)

	return config, nil
}

// GetConfig retrieves a configuration by ID
func (s *ConfigService) GetConfig(ctx context.Context, id uuid.UUID) (*models.Configuration, error) {
	// Try to get from cache first
	config, err := s.getCachedConfig(ctx, id)
	if err == nil {
		return config, nil
	}

	// Get from database
	config, err = s.dbRepo.GetConfiguration(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cacheConfig(ctx, config); err != nil {
		s.logger.Warn("Failed to cache config", zap.Error(err))
	}

	return config, nil
}

// GetConfigs retrieves configurations with pagination
func (s *ConfigService) GetConfigs(ctx context.Context, page, limit int, service, environment string) ([]*models.Configuration, int64, error) {
	offset := (page - 1) * limit
	configs, total, err := s.dbRepo.ListConfigurations(ctx, service, environment, limit, offset)
	return configs, int64(total), err
}

// UpdateConfig updates an existing configuration
func (s *ConfigService) UpdateConfig(ctx context.Context, id uuid.UUID, req *models.UpdateConfigRequest, userID string) (*models.Configuration, error) {
	// Get existing configuration
	config, err := s.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create new version
	newConfig := &models.Configuration{
		ID:          uuid.New(),
		Service:     config.Service,
		Environment: config.Environment,
		Version:     config.Version + 1,
		Data:        req.Data,
		Schema:      req.Schema,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		CreatedBy:   config.CreatedBy,
		UpdatedBy:   uuid.MustParse(userID),
	}

	// Validate the new configuration
	if err := s.validateConfigData(newConfig.Data); err != nil {
		return nil, fmt.Errorf("invalid configuration data: %w", err)
	}

	// Save to database
	if err := s.dbRepo.CreateConfiguration(ctx, newConfig); err != nil {
		s.logger.Error("Failed to create new config version", zap.Error(err))
		return nil, err
	}

	// Update cache
	if err := s.cacheConfig(ctx, newConfig); err != nil {
		s.logger.Warn("Failed to cache updated config", zap.Error(err))
	}

	// Publish configuration change event
	if err := s.publishConfigEvent(ctx, "config.updated", newConfig); err != nil {
		s.logger.Warn("Failed to publish config event", zap.Error(err))
	}

	s.logger.Info("Configuration updated",
		zap.String("config_id", newConfig.ID.String()),
		zap.String("service", newConfig.Service),
		zap.String("environment", newConfig.Environment),
		zap.Int("version", newConfig.Version),
	)

	return newConfig, nil
}

// DeleteConfig deletes a configuration
func (s *ConfigService) DeleteConfig(ctx context.Context, id uuid.UUID, userID string) error {
	// Get configuration first
	config, err := s.GetConfig(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := s.dbRepo.DeleteConfiguration(ctx, id); err != nil {
		s.logger.Error("Failed to delete config from database", zap.Error(err))
		return err
	}

	// Remove from cache
	if err := s.removeCachedConfig(ctx, id); err != nil {
		s.logger.Warn("Failed to remove config from cache", zap.Error(err))
	}

	// Publish configuration change event
	if err := s.publishConfigEvent(ctx, "config.deleted", config); err != nil {
		s.logger.Warn("Failed to publish config event", zap.Error(err))
	}

	s.logger.Info("Configuration deleted",
		zap.String("config_id", id.String()),
		zap.String("service", config.Service),
		zap.String("environment", config.Environment),
	)

	return nil
}

// GetConfigValue retrieves a specific configuration value
func (s *ConfigService) GetConfigValue(ctx context.Context, service, environment, key string) (interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("config:%s:%s:%s", service, environment, key)
	cachedValue, err := s.redisRepo.Get(ctx, cacheKey)
	if err == nil {
		var result interface{}
		if err := json.Unmarshal([]byte(cachedValue), &result); err == nil {
			return result, nil
		}
	}

	// Get from database (0 means latest version)
	config, err := s.dbRepo.GetConfigurationByService(ctx, service, environment, 0)
	if err != nil {
		return nil, err
	}

	// Extract the specific value
	if len(config.Data) == 0 {
		return nil, fmt.Errorf("configuration data is empty")
	}

	// Parse the JSON data
	var dataMap map[string]interface{}
	if err := json.Unmarshal(config.Data, &dataMap); err != nil {
		return nil, fmt.Errorf("failed to parse configuration data: %w", err)
	}

	value, exists := dataMap[key]
	if !exists {
		return nil, fmt.Errorf("configuration key not found: %s", key)
	}

	// Cache the value
	valueBytes, _ := json.Marshal(value)
	if err := s.redisRepo.Set(ctx, cacheKey, string(valueBytes), 300); err != nil {
		s.logger.Warn("Failed to cache config value", zap.Error(err))
	}

	return value, nil
}

// SetConfigValue sets a specific configuration value
func (s *ConfigService) SetConfigValue(ctx context.Context, service, environment, key string, value interface{}, userID string) error {
	// Get latest configuration (0 means latest version)
	config, err := s.dbRepo.GetConfigurationByService(ctx, service, environment, 0)
	if err != nil {
		return err
	}

	// Update the specific value
	var dataMap map[string]interface{}
	if len(config.Data) == 0 {
		dataMap = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(config.Data, &dataMap); err != nil {
			return fmt.Errorf("failed to parse configuration data: %w", err)
		}
	}
	dataMap[key] = value
	
	// Marshal back to JSON
	updatedData, err := json.Marshal(dataMap)
	if err != nil {
		return fmt.Errorf("failed to marshal updated data: %w", err)
	}

	// Create new version
	newConfig := &models.Configuration{
		ID:          uuid.New(),
		Service:     config.Service,
		Environment: config.Environment,
		Type:        config.Type,
		EntityID:    config.EntityID,
		Name:        config.Name,
		Description: config.Description,
		Version:     config.Version + 1,
		Status:      config.Status,
		Data:        updatedData,
		Schema:      config.Schema,
		Tags:        config.Tags,
		Metadata:    config.Metadata,
		CreatedBy:   config.CreatedBy,
		UpdatedBy:   uuid.MustParse(userID),
	}

	// Save to database
	if err := s.dbRepo.CreateConfiguration(ctx, newConfig); err != nil {
		s.logger.Error("Failed to create new config version", zap.Error(err))
		return err
	}

	// Update cache
	if err := s.cacheConfig(ctx, newConfig); err != nil {
		s.logger.Warn("Failed to cache updated config", zap.Error(err))
	}

	// Cache the specific value
	cacheKey := fmt.Sprintf("config:%s:%s:%s", service, environment, key)
	valueBytes, _ := json.Marshal(value)
	if err := s.redisRepo.Set(ctx, cacheKey, string(valueBytes), 300); err != nil {
		s.logger.Warn("Failed to cache config value", zap.Error(err))
	}

	// Publish configuration change event
	if err := s.publishConfigEvent(ctx, "config.value_updated", newConfig); err != nil {
		s.logger.Warn("Failed to publish config event", zap.Error(err))
	}

	return nil
}

// ValidateConfig validates a configuration
func (s *ConfigService) ValidateConfig(ctx context.Context, req *models.ValidateConfigRequest) (*models.ValidationResult, error) {
	result := &models.ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Validate configuration data
	if err := s.validateConfigData(req.Data); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Additional validation logic can be added here

	return result, nil
}

// GetConfigHistory retrieves configuration history
func (s *ConfigService) GetConfigHistory(ctx context.Context, service, environment string, page, limit int) ([]*models.Configuration, int64, error) {
	offset := (page - 1) * limit
	configs, total, err := s.dbRepo.GetConfigHistory(ctx, service, environment, limit, offset)
	return configs, int64(total), err
}

// CreateTemplate creates a new configuration template
func (s *ConfigService) CreateTemplate(ctx context.Context, req *models.CreateTemplateRequest, userID string) (*models.ConfigTemplate, error) {
	template := &models.ConfigTemplate{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Schema:      req.Schema,
		Defaults:    req.Defaults,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		CreatedBy:   uuid.MustParse(userID),
		UpdatedBy:   uuid.MustParse(userID),
	}

	if err := s.dbRepo.CreateTemplate(ctx, template); err != nil {
		s.logger.Error("Failed to create template", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Configuration template created",
		zap.String("template_id", template.ID.String()),
		zap.String("name", template.Name),
	)

	return template, nil
}

// GetTemplates retrieves configuration templates
func (s *ConfigService) GetTemplates(ctx context.Context, page, limit int, category string) ([]*models.ConfigTemplate, int64, error) {
	offset := (page - 1) * limit
	templates, total, err := s.dbRepo.ListTemplates(ctx, category, limit, offset)
	return templates, int64(total), err
}

// ApplyTemplate applies a template to create a configuration
func (s *ConfigService) ApplyTemplate(ctx context.Context, templateID uuid.UUID, req *models.ApplyTemplateRequest, userID string) (*models.Configuration, error) {
	// Get template
	template, err := s.dbRepo.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Merge template defaults with provided variables
	data := make(map[string]interface{})
	
	// Parse template defaults
	if len(template.Defaults) > 0 {
		var defaults map[string]interface{}
		if err := json.Unmarshal(template.Defaults, &defaults); err == nil {
			for k, v := range defaults {
				data[k] = v
			}
		}
	}
	
	// Parse and merge provided variables
	if len(req.Variables) > 0 {
		var variables map[string]interface{}
		if err := json.Unmarshal(req.Variables, &variables); err == nil {
			for k, v := range variables {
				data[k] = v
			}
		}
	}
	
	// Marshal data back to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuration data: %w", err)
	}

	// Create configuration from template
	config := &models.Configuration{
		ID:          uuid.New(),
		Service:     req.Service,
		Environment: req.Environment,
		Type:        template.Type,
		EntityID:    req.EntityID,
		Name:        req.Name,
		Description: req.Description,
		Version:     1,
		Status:      models.ConfigStatusDraft,
		Data:        dataBytes,
		Schema:      template.Schema,
		Tags:        []string{}, // No tags in ApplyTemplateRequest
		Metadata:    make(map[string]interface{}), // Empty metadata
		CreatedBy:   uuid.MustParse(userID),
		UpdatedBy:   uuid.MustParse(userID),
	}

	// Save to database
	if err := s.dbRepo.CreateConfiguration(ctx, config); err != nil {
		s.logger.Error("Failed to create config from template", zap.Error(err))
		return nil, err
	}

	// Cache the configuration
	if err := s.cacheConfig(ctx, config); err != nil {
		s.logger.Warn("Failed to cache config", zap.Error(err))
	}

	// Publish configuration change event
	if err := s.publishConfigEvent(ctx, "config.created_from_template", config); err != nil {
		s.logger.Warn("Failed to publish config event", zap.Error(err))
	}

	s.logger.Info("Configuration created from template",
		zap.String("config_id", config.ID.String()),
		zap.String("template_id", templateID.String()),
		zap.String("service", config.Service),
		zap.String("environment", config.Environment),
	)

	return config, nil
}

// Helper methods

func (s *ConfigService) validateConfigData(data json.RawMessage) error {
	if len(data) == 0 {
		return fmt.Errorf("configuration data cannot be empty")
	}
	
	// Validate that it's valid JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}
	
	// Add more validation logic as needed
	return nil
}

func (s *ConfigService) cacheConfig(ctx context.Context, config *models.Configuration) error {
	cacheKey := fmt.Sprintf("config:%s", config.ID.String())
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return s.redisRepo.Set(ctx, cacheKey, string(configBytes), 3600) // Cache for 1 hour
}

func (s *ConfigService) getCachedConfig(ctx context.Context, id uuid.UUID) (*models.Configuration, error) {
	cacheKey := fmt.Sprintf("config:%s", id.String())
	configData, err := s.redisRepo.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var config models.Configuration
	if err := json.Unmarshal([]byte(configData), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *ConfigService) removeCachedConfig(ctx context.Context, id uuid.UUID) error {
	cacheKey := fmt.Sprintf("config:%s", id.String())
	return s.redisRepo.Delete(ctx, cacheKey)
}

func (s *ConfigService) publishConfigEvent(ctx context.Context, eventType string, config *models.Configuration) error {
	event := map[string]interface{}{
		"type":        eventType,
		"config_id":   config.ID.String(),
		"service":     config.Service,
		"environment": config.Environment,
		"version":     config.Version,
		"timestamp":   config.UpdatedAt,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.natsRepo.Publish(ctx, "config.events", eventBytes)
}