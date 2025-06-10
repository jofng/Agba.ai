package service

import (
	"context"
	"fmt"
	"time"

	"freeswitch-service/internal/config"
	"freeswitch-service/internal/freeswitch"
	"freeswitch-service/internal/integration"
	"freeswitch-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FreeSWITCHService handles business logic for FreeSWITCH operations
type FreeSWITCHService struct {
	config      *config.Config
	fsManager   *freeswitch.Manager
	integration *integration.ServiceIntegration
	logger      *zap.Logger
}

// NewFreeSWITCHService creates a new FreeSWITCH service
func NewFreeSWITCHService(
	config *config.Config,
	fsManager *freeswitch.Manager,
	integration *integration.ServiceIntegration,
	logger *zap.Logger,
) *FreeSWITCHService {
	service := &FreeSWITCHService{
		config:      config,
		fsManager:   fsManager,
		integration: integration,
		logger:      logger,
	}

	// Register event handlers
	service.registerEventHandlers()

	return service
}

// Start starts the FreeSWITCH service
func (s *FreeSWITCHService) Start(ctx context.Context) error {
	s.logger.Info("Starting FreeSWITCH service")

	// Connect to FreeSWITCH
	if err := s.fsManager.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to FreeSWITCH: %w", err)
	}

	s.logger.Info("FreeSWITCH service started successfully")
	return nil
}

// Stop stops the FreeSWITCH service
func (s *FreeSWITCHService) Stop() error {
	s.logger.Info("Stopping FreeSWITCH service")

	if err := s.fsManager.Disconnect(); err != nil {
		s.logger.Error("Failed to disconnect from FreeSWITCH", zap.Error(err))
	}

	s.logger.Info("FreeSWITCH service stopped")
	return nil
}

// OriginateCall initiates an outbound call
func (s *FreeSWITCHService) OriginateCall(ctx context.Context, request *models.CallRequest) (*models.CallResponse, error) {
	s.logger.Info("Originating call",
		zap.String("caller_id", request.CallerID),
		zap.String("called_number", request.CalledNumber))

	// Generate call ID
	callID := uuid.New()

	// Originate call through FreeSWITCH
	sessionID, err := s.fsManager.OriginateCall(request.CallerID, request.CalledNumber, "default")
	if err != nil {
		s.logger.Error("Failed to originate call", zap.Error(err))
		return nil, fmt.Errorf("failed to originate call: %w", err)
	}

	// Create call record
	call := &models.Call{
		ID:           callID,
		SessionID:    sessionID,
		CallerID:     request.CallerID,
		CalledNumber: request.CalledNumber,
		Direction:    models.CallDirectionOutbound,
		Status:       models.CallStatusRinging,
		StartTime:    time.Now(),
		UserID:       request.UserID,
		TrunkID:      request.TrunkID,
		Metadata:     request.Variables,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Store call in database through integration
	if err := s.integration.StoreCall(ctx, call); err != nil {
		s.logger.Error("Failed to store call", zap.Error(err))
		// Don't fail the call, just log the error
	}

	// Start recording if requested
	if request.Recording {
		if err := s.startCallRecording(ctx, call); err != nil {
			s.logger.Error("Failed to start recording", zap.Error(err))
			// Don't fail the call, just log the error
		}
	}

	// Send metrics
	s.integration.SendCallMetrics(ctx, call, "call_originated")

	response := &models.CallResponse{
		CallID:    callID,
		SessionID: sessionID,
		Status:    "initiated",
		Message:   "Call originated successfully",
	}

	return response, nil
}

// HangupCall hangs up an active call
func (s *FreeSWITCHService) HangupCall(ctx context.Context, callID uuid.UUID) error {
	s.logger.Info("Hanging up call", zap.String("call_id", callID.String()))

	// Get call from database
	call, err := s.integration.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}

	// Hangup call through FreeSWITCH
	if err := s.fsManager.HangupCall(call.SessionID); err != nil {
		return fmt.Errorf("failed to hangup call: %w", err)
	}

	// Update call status
	call.Status = models.CallStatusHangup
	now := time.Now()
	call.EndTime = &now
	call.UpdatedAt = now

	if call.AnswerTime != nil {
		call.Duration = int(now.Sub(*call.AnswerTime).Seconds())
	}

	// Update call in database
	if err := s.integration.UpdateCall(ctx, call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	// Send metrics
	s.integration.SendCallMetrics(ctx, call, "call_hangup")

	return nil
}

// TransferCall transfers a call to a new destination
func (s *FreeSWITCHService) TransferCall(ctx context.Context, request *models.CallTransferRequest) error {
	s.logger.Info("Transferring call",
		zap.String("call_id", request.CallID.String()),
		zap.String("destination", request.Destination))

	// Get call from database
	call, err := s.integration.GetCall(ctx, request.CallID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}

	// Transfer call through FreeSWITCH
	if err := s.fsManager.TransferCall(call.SessionID, request.Destination); err != nil {
		return fmt.Errorf("failed to transfer call: %w", err)
	}

	// Update call status
	call.Status = models.CallStatusTransfer
	call.UpdatedAt = time.Now()

	// Update call in database
	if err := s.integration.UpdateCall(ctx, call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	// Send metrics
	s.integration.SendCallMetrics(ctx, call, "call_transfer")

	return nil
}

// HoldCall puts a call on hold or removes from hold
func (s *FreeSWITCHService) HoldCall(ctx context.Context, request *models.CallHoldRequest) error {
	s.logger.Info("Hold/Unhold call",
		zap.String("call_id", request.CallID.String()),
		zap.Bool("hold", request.Hold))

	// Get call from database
	call, err := s.integration.GetCall(ctx, request.CallID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}

	// Hold/Unhold call through FreeSWITCH
	if request.Hold {
		if err := s.fsManager.HoldCall(call.SessionID); err != nil {
			return fmt.Errorf("failed to hold call: %w", err)
		}
		call.Status = models.CallStatusHold
	} else {
		if err := s.fsManager.UnholdCall(call.SessionID); err != nil {
			return fmt.Errorf("failed to unhold call: %w", err)
		}
		call.Status = models.CallStatusBridged
	}

	// Update call status
	call.UpdatedAt = time.Now()

	// Update call in database
	if err := s.integration.UpdateCall(ctx, call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	// Send metrics
	action := "call_hold"
	if !request.Hold {
		action = "call_unhold"
	}
	s.integration.SendCallMetrics(ctx, call, action)

	return nil
}

// ControlRecording controls call recording (start, stop, pause, resume)
func (s *FreeSWITCHService) ControlRecording(ctx context.Context, request *models.CallRecordingRequest) error {
	s.logger.Info("Control recording",
		zap.String("call_id", request.CallID.String()),
		zap.String("action", request.Action))

	// Get call from database
	call, err := s.integration.GetCall(ctx, request.CallID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}

	switch request.Action {
	case "start":
		return s.startCallRecording(ctx, call)
	case "stop":
		return s.stopCallRecording(ctx, call)
	case "pause":
		// Implementation for pause recording
		return s.pauseCallRecording(ctx, call)
	case "resume":
		// Implementation for resume recording
		return s.resumeCallRecording(ctx, call)
	default:
		return fmt.Errorf("invalid recording action: %s", request.Action)
	}
}

// GetCall retrieves call information
func (s *FreeSWITCHService) GetCall(ctx context.Context, callID uuid.UUID) (*models.Call, error) {
	return s.integration.GetCall(ctx, callID)
}

// GetActiveCalls retrieves all active calls
func (s *FreeSWITCHService) GetActiveCalls(ctx context.Context) ([]*models.Call, error) {
	// Get active calls from FreeSWITCH manager
	callStates := s.fsManager.GetCallStates()

	var activeCalls []*models.Call
	for _, call := range callStates {
		activeCalls = append(activeCalls, call)
	}

	return activeCalls, nil
}

// GetCallStats retrieves call statistics
func (s *FreeSWITCHService) GetCallStats(ctx context.Context) (*models.CallStats, error) {
	stats, err := s.integration.GetCallStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get call stats: %w", err)
	}

	// Add real-time data from FreeSWITCH
	stats.ActiveCalls = int64(s.fsManager.GetActiveCalls())
	stats.LastUpdated = time.Now()

	return stats, nil
}

// HealthCheck performs health checks
func (s *FreeSWITCHService) HealthCheck(ctx context.Context) (bool, map[string]interface{}) {
	checks := make(map[string]interface{})
	healthy := true

	// Check FreeSWITCH connection
	if s.fsManager.IsConnected() {
		checks["freeswitch"] = map[string]interface{}{
			"status":  "healthy",
			"message": "Connected to FreeSWITCH Event Socket",
		}
	} else {
		checks["freeswitch"] = map[string]interface{}{
			"status":  "unhealthy",
			"message": "Not connected to FreeSWITCH Event Socket",
		}
		healthy = false
	}

	// Check service integrations
	integrationHealth := s.integration.HealthCheck(ctx)
	for service, status := range integrationHealth {
		checks[service] = status
		if statusMap, ok := status.(map[string]interface{}); ok {
			if statusMap["status"] != "healthy" {
				healthy = false
			}
		}
	}

	// Check active calls
	activeCalls := s.fsManager.GetActiveCalls()
	checks["active_calls"] = map[string]interface{}{
		"count":  activeCalls,
		"status": "healthy",
	}

	return healthy, checks
}

// registerEventHandlers registers FreeSWITCH event handlers
func (s *FreeSWITCHService) registerEventHandlers() {
	// Handle channel create events
	s.fsManager.RegisterEventHandler("CHANNEL_CREATE", freeswitch.EventHandlerFunc(func(event *models.FreeSWITCHEvent) error {
		s.logger.Debug("Channel created", zap.String("session_id", event.SessionID))
		return s.handleChannelCreate(event)
	}))

	// Handle channel answer events
	s.fsManager.RegisterEventHandler("CHANNEL_ANSWER", freeswitch.EventHandlerFunc(func(event *models.FreeSWITCHEvent) error {
		s.logger.Debug("Channel answered", zap.String("session_id", event.SessionID))
		return s.handleChannelAnswer(event)
	}))

	// Handle channel hangup events
	s.fsManager.RegisterEventHandler("CHANNEL_HANGUP", freeswitch.EventHandlerFunc(func(event *models.FreeSWITCHEvent) error {
		s.logger.Debug("Channel hangup", zap.String("session_id", event.SessionID))
		return s.handleChannelHangup(event)
	}))

	// Handle channel bridge events
	s.fsManager.RegisterEventHandler("CHANNEL_BRIDGE", freeswitch.EventHandlerFunc(func(event *models.FreeSWITCHEvent) error {
		s.logger.Debug("Channel bridged", zap.String("session_id", event.SessionID))
		return s.handleChannelBridge(event)
	}))

	// Handle recording events
	s.fsManager.RegisterEventHandler("RECORD_START", freeswitch.EventHandlerFunc(func(event *models.FreeSWITCHEvent) error {
		s.logger.Debug("Recording started", zap.String("session_id", event.SessionID))
		return s.handleRecordingStart(event)
	}))

	s.fsManager.RegisterEventHandler("RECORD_STOP", freeswitch.EventHandlerFunc(func(event *models.FreeSWITCHEvent) error {
		s.logger.Debug("Recording stopped", zap.String("session_id", event.SessionID))
		return s.handleRecordingStop(event)
	}))
}

// Event handlers
func (s *FreeSWITCHService) handleChannelCreate(event *models.FreeSWITCHEvent) error {
	// Handle inbound calls
	if event.CallerID != "" && event.Destination != "" {
		call := &models.Call{
			ID:           uuid.New(),
			SessionID:    event.SessionID,
			CallerID:     event.CallerID,
			CalledNumber: event.Destination,
			Direction:    models.CallDirectionInbound,
			Status:       models.CallStatusRinging,
			StartTime:    event.Timestamp,
			Metadata:     event.Variables,
			CreatedAt:    event.Timestamp,
			UpdatedAt:    event.Timestamp,
		}

		// Store call in database
		if err := s.integration.StoreCall(context.Background(), call); err != nil {
			s.logger.Error("Failed to store inbound call", zap.Error(err))
		}

		// Send metrics
		s.integration.SendCallMetrics(context.Background(), call, "call_created")
	}

	return nil
}

func (s *FreeSWITCHService) handleChannelAnswer(event *models.FreeSWITCHEvent) error {
	// Update call status to answered
	call, err := s.integration.GetCallBySessionID(context.Background(), event.SessionID)
	if err != nil {
		s.logger.Error("Failed to get call by session ID", zap.Error(err))
		return err
	}

	call.Status = models.CallStatusAnswered
	now := event.Timestamp
	call.AnswerTime = &now
	call.UpdatedAt = now

	if err := s.integration.UpdateCall(context.Background(), call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	// Send metrics
	s.integration.SendCallMetrics(context.Background(), call, "call_answered")

	return nil
}

func (s *FreeSWITCHService) handleChannelHangup(event *models.FreeSWITCHEvent) error {
	// Update call status to hangup
	call, err := s.integration.GetCallBySessionID(context.Background(), event.SessionID)
	if err != nil {
		s.logger.Error("Failed to get call by session ID", zap.Error(err))
		return err
	}

	call.Status = models.CallStatusHangup
	call.HangupCause = event.HangupCause
	now := event.Timestamp
	call.EndTime = &now
	call.UpdatedAt = now

	if call.AnswerTime != nil {
		call.Duration = int(now.Sub(*call.AnswerTime).Seconds())
	}

	if err := s.integration.UpdateCall(context.Background(), call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	// Send metrics
	s.integration.SendCallMetrics(context.Background(), call, "call_hangup")

	return nil
}

func (s *FreeSWITCHService) handleChannelBridge(event *models.FreeSWITCHEvent) error {
	// Update call status to bridged
	call, err := s.integration.GetCallBySessionID(context.Background(), event.SessionID)
	if err != nil {
		s.logger.Error("Failed to get call by session ID", zap.Error(err))
		return err
	}

	call.Status = models.CallStatusBridged
	call.UpdatedAt = event.Timestamp

	if err := s.integration.UpdateCall(context.Background(), call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	// Send metrics
	s.integration.SendCallMetrics(context.Background(), call, "call_bridged")

	return nil
}

func (s *FreeSWITCHService) handleRecordingStart(event *models.FreeSWITCHEvent) error {
	// Update call recording status
	call, err := s.integration.GetCallBySessionID(context.Background(), event.SessionID)
	if err != nil {
		s.logger.Error("Failed to get call by session ID", zap.Error(err))
		return err
	}

	call.Status = models.CallStatusRecording
	call.UpdatedAt = event.Timestamp

	if err := s.integration.UpdateCall(context.Background(), call); err != nil {
		s.logger.Error("Failed to update call", zap.Error(err))
	}

	return nil
}

func (s *FreeSWITCHService) handleRecordingStop(event *models.FreeSWITCHEvent) error {
	// Handle recording completion
	call, err := s.integration.GetCallBySessionID(context.Background(), event.SessionID)
	if err != nil {
		s.logger.Error("Failed to get call by session ID", zap.Error(err))
		return err
	}

	// Notify recording service about completion
	if call.RecordingID != nil {
		if err := s.integration.CompleteRecording(context.Background(), *call.RecordingID); err != nil {
			s.logger.Error("Failed to complete recording", zap.Error(err))
		}
	}

	return nil
}

// Recording helper methods
func (s *FreeSWITCHService) startCallRecording(ctx context.Context, call *models.Call) error {
	// Create recording through integration
	recordingID, filename, err := s.integration.StartRecording(ctx, call)
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Start recording in FreeSWITCH
	if err := s.fsManager.StartRecording(call.SessionID, filename); err != nil {
		return fmt.Errorf("failed to start FreeSWITCH recording: %w", err)
	}

	// Update call with recording ID
	call.RecordingID = &recordingID
	call.UpdatedAt = time.Now()

	if err := s.integration.UpdateCall(ctx, call); err != nil {
		s.logger.Error("Failed to update call with recording ID", zap.Error(err))
	}

	return nil
}

func (s *FreeSWITCHService) stopCallRecording(ctx context.Context, call *models.Call) error {
	if call.RecordingID == nil {
		return fmt.Errorf("no active recording for call")
	}

	// Get recording filename
	filename, err := s.integration.GetRecordingFilename(ctx, *call.RecordingID)
	if err != nil {
		return fmt.Errorf("failed to get recording filename: %w", err)
	}

	// Stop recording in FreeSWITCH
	if err := s.fsManager.StopRecording(call.SessionID, filename); err != nil {
		return fmt.Errorf("failed to stop FreeSWITCH recording: %w", err)
	}

	// Update recording status
	if err := s.integration.StopRecording(ctx, *call.RecordingID); err != nil {
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	return nil
}

func (s *FreeSWITCHService) pauseCallRecording(ctx context.Context, call *models.Call) error {
	// Implementation for pause recording
	// This would require additional FreeSWITCH commands
	return fmt.Errorf("pause recording not implemented")
}

func (s *FreeSWITCHService) resumeCallRecording(ctx context.Context, call *models.Call) error {
	// Implementation for resume recording
	// This would require additional FreeSWITCH commands
	return fmt.Errorf("resume recording not implemented")
}