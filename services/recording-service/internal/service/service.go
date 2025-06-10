package service

import (
    "context"
    "fmt"
    "time"

    "recording-service/internal/repository"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

// Service handles business logic for recording operations
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

// StartRecording starts a new recording session
func (s *Service) StartRecording(ctx context.Context, sessionID string, userID *uuid.UUID) (*repository.Recording, error) {
    s.logger.Info("Starting recording session", zap.String("session_id", sessionID))

    recording := &repository.Recording{
        ID:        uuid.New(),
        SessionID: sessionID,
        UserID:    userID,
        Filename:  fmt.Sprintf("recording_%s_%d.wav", sessionID, time.Now().Unix()),
        FilePath:  fmt.Sprintf("/recordings/%s/", sessionID),
        Format:    "wav",
        Quality:   "high",
        Status:    "recording",
        Metadata:  make(map[string]interface{}),
        StartedAt: time.Now(),
    }

    if err := s.repository.CreateRecording(ctx, recording); err != nil {
        return nil, fmt.Errorf("failed to create recording: %w", err)
    }

    return recording, nil
}

// StopRecording stops a recording session
func (s *Service) StopRecording(ctx context.Context, id uuid.UUID) error {
    s.logger.Info("Stopping recording", zap.String("id", id.String()))

    return s.repository.UpdateRecordingStatus(ctx, id, "completed")
}

// GetRecording retrieves a recording by ID
func (s *Service) GetRecording(ctx context.Context, id uuid.UUID) (*repository.Recording, error) {
    return s.repository.GetRecording(ctx, id)
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
