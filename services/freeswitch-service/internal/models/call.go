package models

import (
	"time"

	"github.com/google/uuid"
)

// CallDirection represents the direction of a call
type CallDirection string

const (
	CallDirectionInbound  CallDirection = "inbound"
	CallDirectionOutbound CallDirection = "outbound"
)

// CallStatus represents the current status of a call
type CallStatus string

const (
	CallStatusRinging    CallStatus = "ringing"
	CallStatusAnswered   CallStatus = "answered"
	CallStatusBridged    CallStatus = "bridged"
	CallStatusHangup     CallStatus = "hangup"
	CallStatusFailed     CallStatus = "failed"
	CallStatusTransfer   CallStatus = "transfer"
	CallStatusHold       CallStatus = "hold"
	CallStatusRecording  CallStatus = "recording"
)

// Call represents a phone call in the system
type Call struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	SessionID         string                 `json:"session_id" db:"session_id"`
	CallerID          string                 `json:"caller_id" db:"caller_id"`
	CalledNumber      string                 `json:"called_number" db:"called_number"`
	Direction         CallDirection          `json:"direction" db:"direction"`
	Status            CallStatus             `json:"status" db:"status"`
	StartTime         time.Time              `json:"start_time" db:"start_time"`
	AnswerTime        *time.Time             `json:"answer_time,omitempty" db:"answer_time"`
	EndTime           *time.Time             `json:"end_time,omitempty" db:"end_time"`
	Duration          int                    `json:"duration" db:"duration"` // in seconds
	RecordingID       *uuid.UUID             `json:"recording_id,omitempty" db:"recording_id"`
	UserID            *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	TrunkID           string                 `json:"trunk_id" db:"trunk_id"`
	HangupCause       string                 `json:"hangup_cause,omitempty" db:"hangup_cause"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// CallEvent represents a call event for real-time updates
type CallEvent struct {
	CallID    uuid.UUID              `json:"call_id"`
	EventType string                 `json:"event_type"`
	Status    CallStatus             `json:"status"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// CallRequest represents a request to initiate a call
type CallRequest struct {
	CallerID     string                 `json:"caller_id" binding:"required"`
	CalledNumber string                 `json:"called_number" binding:"required"`
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	TrunkID      string                 `json:"trunk_id"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Recording    bool                   `json:"recording,omitempty"`
}

// CallResponse represents the response after initiating a call
type CallResponse struct {
	CallID    uuid.UUID `json:"call_id"`
	SessionID string    `json:"session_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
}

// CallStats represents call statistics
type CallStats struct {
	TotalCalls       int64   `json:"total_calls"`
	ActiveCalls      int64   `json:"active_calls"`
	SuccessfulCalls  int64   `json:"successful_calls"`
	FailedCalls      int64   `json:"failed_calls"`
	AverageCallTime  float64 `json:"average_call_time"`
	CallSuccessRate  float64 `json:"call_success_rate"`
	PeakConcurrency  int64   `json:"peak_concurrency"`
	LastUpdated      time.Time `json:"last_updated"`
}

// SIPTrunk represents a SIP trunk configuration
type SIPTrunk struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Host        string                 `json:"host" db:"host"`
	Port        int                    `json:"port" db:"port"`
	Username    string                 `json:"username" db:"username"`
	Password    string                 `json:"password" db:"password"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	MaxChannels int                    `json:"max_channels" db:"max_channels"`
	Priority    int                    `json:"priority" db:"priority"`
	Config      map[string]interface{} `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// FreeSWITCHEvent represents an event from FreeSWITCH
type FreeSWITCHEvent struct {
	EventName    string                 `json:"event_name"`
	CoreUUID     string                 `json:"core_uuid"`
	SessionID    string                 `json:"session_id,omitempty"`
	CallerID     string                 `json:"caller_id,omitempty"`
	Destination  string                 `json:"destination,omitempty"`
	CallState    string                 `json:"call_state,omitempty"`
	HangupCause  string                 `json:"hangup_cause,omitempty"`
	Variables    map[string]interface{} `json:"variables"`
	Timestamp    time.Time              `json:"timestamp"`
}

// CallTransferRequest represents a call transfer request
type CallTransferRequest struct {
	CallID      uuid.UUID `json:"call_id" binding:"required"`
	Destination string    `json:"destination" binding:"required"`
	Type        string    `json:"type"` // "blind" or "attended"
}

// CallHoldRequest represents a call hold/unhold request
type CallHoldRequest struct {
	CallID uuid.UUID `json:"call_id" binding:"required"`
	Hold   bool      `json:"hold"`
}

// CallRecordingRequest represents a call recording control request
type CallRecordingRequest struct {
	CallID uuid.UUID `json:"call_id" binding:"required"`
	Action string    `json:"action"` // "start", "stop", "pause", "resume"
}