package freeswitch

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"freeswitch-service/internal/config"
	"freeswitch-service/internal/models"
	"go.uber.org/zap"
)

// Manager handles FreeSWITCH Event Socket connections and operations
type Manager struct {
	config     *config.FreeSWITCHConfig
	logger     *zap.Logger
	conn       net.Conn
	connected  bool
	mu         sync.RWMutex
	eventChan  chan *models.FreeSWITCHEvent
	stopChan   chan struct{}
	handlers   map[string]EventHandler
	callStates map[string]*models.Call
}

// EventHandler defines the interface for handling FreeSWITCH events
type EventHandler interface {
	HandleEvent(event *models.FreeSWITCHEvent) error
}

// EventHandlerFunc is a function type that implements EventHandler
type EventHandlerFunc func(event *models.FreeSWITCHEvent) error

// HandleEvent implements EventHandler interface
func (f EventHandlerFunc) HandleEvent(event *models.FreeSWITCHEvent) error {
	return f(event)
}

// NewManager creates a new FreeSWITCH manager
func NewManager(config *config.FreeSWITCHConfig, logger *zap.Logger) *Manager {
	return &Manager{
		config:     config,
		logger:     logger,
		eventChan:  make(chan *models.FreeSWITCHEvent, 1000),
		stopChan:   make(chan struct{}),
		handlers:   make(map[string]EventHandler),
		callStates: make(map[string]*models.Call),
	}
}

// Connect establishes connection to FreeSWITCH Event Socket
func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", m.config.EventSocketHost, m.config.EventSocketPort)
	
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to FreeSWITCH: %w", err)
	}

	m.conn = conn
	m.connected = true

	// Authenticate
	if err := m.authenticate(); err != nil {
		m.conn.Close()
		m.connected = false
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Subscribe to events
	if err := m.subscribeToEvents(); err != nil {
		m.conn.Close()
		m.connected = false
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Start event processing
	go m.processEvents(ctx)
	go m.readEvents(ctx)

	m.logger.Info("Connected to FreeSWITCH Event Socket", zap.String("address", addr))
	return nil
}

// Disconnect closes the connection to FreeSWITCH
func (m *Manager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil
	}

	close(m.stopChan)
	
	if m.conn != nil {
		m.conn.Close()
	}

	m.connected = false
	m.logger.Info("Disconnected from FreeSWITCH Event Socket")
	return nil
}

// authenticate authenticates with FreeSWITCH Event Socket
func (m *Manager) authenticate() error {
	// Read the initial auth request
	reader := bufio.NewReader(m.conn)
	_, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read auth request: %w", err)
	}

	// Send authentication
	authCmd := fmt.Sprintf("auth %s\n\n", m.config.Password)
	_, err = m.conn.Write([]byte(authCmd))
	if err != nil {
		return fmt.Errorf("failed to send auth command: %w", err)
	}

	// Read auth response
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("authentication failed: %s", response)
	}

	return nil
}

// subscribeToEvents subscribes to FreeSWITCH events
func (m *Manager) subscribeToEvents() error {
	events := []string{
		"CHANNEL_CREATE",
		"CHANNEL_ANSWER",
		"CHANNEL_HANGUP",
		"CHANNEL_BRIDGE",
		"CHANNEL_UNBRIDGE",
		"CHANNEL_PROGRESS",
		"CHANNEL_PROGRESS_MEDIA",
		"DTMF",
		"RECORD_START",
		"RECORD_STOP",
		"CUSTOM",
	}

	for _, event := range events {
		cmd := fmt.Sprintf("event plain %s\n\n", event)
		_, err := m.conn.Write([]byte(cmd))
		if err != nil {
			return fmt.Errorf("failed to subscribe to event %s: %w", event, err)
		}
	}

	return nil
}

// readEvents reads events from FreeSWITCH Event Socket
func (m *Manager) readEvents(ctx context.Context) {
	reader := bufio.NewReader(m.conn)
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		default:
			// Read event data
			eventData, err := m.readEventData(reader)
			if err != nil {
				m.logger.Error("Failed to read event data", zap.Error(err))
				continue
			}

			// Parse event
			event := m.parseEvent(eventData)
			if event != nil {
				select {
				case m.eventChan <- event:
				case <-ctx.Done():
					return
				case <-m.stopChan:
					return
				default:
					m.logger.Warn("Event channel full, dropping event")
				}
			}
		}
	}
}

// readEventData reads a complete event from the socket
func (m *Manager) readEventData(reader *bufio.Reader) (map[string]string, error) {
	eventData := make(map[string]string)
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of event
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			eventData[key] = value
		}
	}

	return eventData, nil
}

// parseEvent parses raw event data into a FreeSWITCHEvent
func (m *Manager) parseEvent(eventData map[string]string) *models.FreeSWITCHEvent {
	eventName, exists := eventData["Event-Name"]
	if !exists {
		return nil
	}

	event := &models.FreeSWITCHEvent{
		EventName:   eventName,
		CoreUUID:    eventData["Core-UUID"],
		SessionID:   eventData["Unique-ID"],
		CallerID:    eventData["Caller-Caller-ID-Number"],
		Destination: eventData["Caller-Destination-Number"],
		CallState:   eventData["Channel-Call-State"],
		HangupCause: eventData["Hangup-Cause"],
		Variables:   make(map[string]interface{}),
		Timestamp:   time.Now(),
	}

	// Copy all event data as variables
	for key, value := range eventData {
		event.Variables[key] = value
	}

	return event
}

// processEvents processes events from the event channel
func (m *Manager) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case event := <-m.eventChan:
			m.handleEvent(event)
		}
	}
}

// handleEvent handles a FreeSWITCH event
func (m *Manager) handleEvent(event *models.FreeSWITCHEvent) {
	m.logger.Debug("Received FreeSWITCH event",
		zap.String("event_name", event.EventName),
		zap.String("session_id", event.SessionID),
		zap.String("caller_id", event.CallerID))

	// Update call state
	m.updateCallState(event)

	// Call registered handlers
	if handler, exists := m.handlers[event.EventName]; exists {
		if err := handler.HandleEvent(event); err != nil {
			m.logger.Error("Event handler failed",
				zap.String("event_name", event.EventName),
				zap.Error(err))
		}
	}

	// Call global handler if exists
	if handler, exists := m.handlers["*"]; exists {
		if err := handler.HandleEvent(event); err != nil {
			m.logger.Error("Global event handler failed", zap.Error(err))
		}
	}
}

// updateCallState updates the internal call state based on events
func (m *Manager) updateCallState(event *models.FreeSWITCHEvent) {
	if event.SessionID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	call, exists := m.callStates[event.SessionID]
	if !exists && event.EventName == "CHANNEL_CREATE" {
		// Create new call state
		call = &models.Call{
			SessionID:    event.SessionID,
			CallerID:     event.CallerID,
			CalledNumber: event.Destination,
			Status:       models.CallStatusRinging,
			StartTime:    event.Timestamp,
			Metadata:     make(map[string]interface{}),
		}
		m.callStates[event.SessionID] = call
	}

	if call == nil {
		return
	}

	// Update call status based on event
	switch event.EventName {
	case "CHANNEL_ANSWER":
		call.Status = models.CallStatusAnswered
		now := event.Timestamp
		call.AnswerTime = &now
	case "CHANNEL_BRIDGE":
		call.Status = models.CallStatusBridged
	case "CHANNEL_HANGUP":
		call.Status = models.CallStatusHangup
		call.HangupCause = event.HangupCause
		now := event.Timestamp
		call.EndTime = &now
		if call.AnswerTime != nil {
			call.Duration = int(now.Sub(*call.AnswerTime).Seconds())
		}
		// Remove from active calls
		delete(m.callStates, event.SessionID)
	}

	// Update metadata
	for key, value := range event.Variables {
		call.Metadata[key] = value
	}
}

// RegisterEventHandler registers an event handler for specific events
func (m *Manager) RegisterEventHandler(eventName string, handler EventHandler) {
	m.handlers[eventName] = handler
}

// SendCommand sends a command to FreeSWITCH
func (m *Manager) SendCommand(command string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return "", fmt.Errorf("not connected to FreeSWITCH")
	}

	// Send command
	cmd := fmt.Sprintf("api %s\n\n", command)
	_, err := m.conn.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	reader := bufio.NewReader(m.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// OriginateCall initiates an outbound call
func (m *Manager) OriginateCall(callerID, destination, context string) (string, error) {
	command := fmt.Sprintf("originate {origination_caller_id_number=%s}sofia/gateway/%s/%s &echo",
		callerID, m.config.DefaultTrunk, destination)
	
	response, err := m.SendCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to originate call: %w", err)
	}

	// Parse session ID from response
	if strings.Contains(response, "+OK") {
		parts := strings.Fields(response)
		if len(parts) > 1 {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("call origination failed: %s", response)
}

// HangupCall hangs up a call
func (m *Manager) HangupCall(sessionID string) error {
	command := fmt.Sprintf("uuid_kill %s", sessionID)
	response, err := m.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to hangup call: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("hangup failed: %s", response)
	}

	return nil
}

// TransferCall transfers a call to a new destination
func (m *Manager) TransferCall(sessionID, destination string) error {
	command := fmt.Sprintf("uuid_transfer %s %s", sessionID, destination)
	response, err := m.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to transfer call: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("transfer failed: %s", response)
	}

	return nil
}

// HoldCall puts a call on hold
func (m *Manager) HoldCall(sessionID string) error {
	command := fmt.Sprintf("uuid_hold %s", sessionID)
	response, err := m.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to hold call: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("hold failed: %s", response)
	}

	return nil
}

// UnholdCall removes a call from hold
func (m *Manager) UnholdCall(sessionID string) error {
	command := fmt.Sprintf("uuid_hold off %s", sessionID)
	response, err := m.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to unhold call: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("unhold failed: %s", response)
	}

	return nil
}

// StartRecording starts recording a call
func (m *Manager) StartRecording(sessionID, filename string) error {
	command := fmt.Sprintf("uuid_record %s start %s", sessionID, filename)
	response, err := m.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("start recording failed: %s", response)
	}

	return nil
}

// StopRecording stops recording a call
func (m *Manager) StopRecording(sessionID, filename string) error {
	command := fmt.Sprintf("uuid_record %s stop %s", sessionID, filename)
	response, err := m.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	if !strings.Contains(response, "+OK") {
		return fmt.Errorf("stop recording failed: %s", response)
	}

	return nil
}

// GetCallStates returns current call states
func (m *Manager) GetCallStates() map[string]*models.Call {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	states := make(map[string]*models.Call)
	for k, v := range m.callStates {
		states[k] = v
	}

	return states
}

// GetActiveCalls returns the number of active calls
func (m *Manager) GetActiveCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.callStates)
}

// IsConnected returns whether the manager is connected to FreeSWITCH
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.connected
}