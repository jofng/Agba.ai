package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RecordingStatus represents the status of a recording
type RecordingStatus string

const (
	RecordingStatusPending     RecordingStatus = "pending"
	RecordingStatusRecording   RecordingStatus = "recording"
	RecordingStatusProcessing  RecordingStatus = "processing"
	RecordingStatusCompleted   RecordingStatus = "completed"
	RecordingStatusFailed      RecordingStatus = "failed"
	RecordingStatusArchived    RecordingStatus = "archived"
	RecordingStatusDeleted     RecordingStatus = "deleted"
	RecordingStatusPaused      RecordingStatus = "paused"
	RecordingStatusStopped     RecordingStatus = "stopped"
)

// RecordingType represents the type of recording
type RecordingType string

const (
	RecordingTypeCall      RecordingType = "call"
	RecordingTypeConference RecordingType = "conference"
	RecordingTypeVoicemail RecordingType = "voicemail"
	RecordingTypeUpload    RecordingType = "upload"
	RecordingTypeStream    RecordingType = "stream"
)

// AudioFormat represents the audio format
type AudioFormat string

const (
	AudioFormatWAV  AudioFormat = "wav"
	AudioFormatMP3  AudioFormat = "mp3"
	AudioFormatFLAC AudioFormat = "flac"
	AudioFormatOGG  AudioFormat = "ogg"
	AudioFormatM4A  AudioFormat = "m4a"
	AudioFormatOPUS AudioFormat = "opus"
	AudioFormatAAC  AudioFormat = "aac"
)

// StorageProvider represents the storage provider
type StorageProvider string

const (
	StorageProviderFile StorageProvider = "file"
	StorageProviderS3   StorageProvider = "s3"
	StorageProviderGCS  StorageProvider = "gcs"
)

// ProcessingStatus represents the status of processing
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusInProgress ProcessingStatus = "in_progress"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
	ProcessingStatusCancelled  ProcessingStatus = "cancelled"
)

// Recording represents a recording in the system
type Recording struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	SessionID   *uuid.UUID                `json:"session_id,omitempty" db:"session_id"`
	CallID      *uuid.UUID                `json:"call_id,omitempty" db:"call_id"`
	UserID      *uuid.UUID                `json:"user_id,omitempty" db:"user_id"`
	
	// Basic information
	Name        string                    `json:"name" db:"name"`
	Description string                    `json:"description,omitempty" db:"description"`
	Type        RecordingType             `json:"type" db:"type"`
	Status      RecordingStatus           `json:"status" db:"status"`
	
	// Audio properties
	Format      AudioFormat               `json:"format" db:"format"`
	SampleRate  int                       `json:"sample_rate" db:"sample_rate"`
	Channels    int                       `json:"channels" db:"channels"`
	BitDepth    int                       `json:"bit_depth" db:"bit_depth"`
	Duration    time.Duration             `json:"duration" db:"duration"`
	FileSize    int64                     `json:"file_size" db:"file_size"`
	
	// Storage information
	StorageProvider StorageProvider       `json:"storage_provider" db:"storage_provider"`
	StoragePath     string                `json:"storage_path" db:"storage_path"`
	StorageBucket   string                `json:"storage_bucket,omitempty" db:"storage_bucket"`
	StorageRegion   string                `json:"storage_region,omitempty" db:"storage_region"`
	
	// Encryption
	IsEncrypted     bool                  `json:"is_encrypted" db:"is_encrypted"`
	EncryptionKey   string                `json:"-" db:"encryption_key"`
	EncryptionIV    string                `json:"-" db:"encryption_iv"`
	
	// Processing information
	ProcessingStatus ProcessingStatus     `json:"processing_status" db:"processing_status"`
	ProcessingStartedAt *time.Time        `json:"processing_started_at,omitempty" db:"processing_started_at"`
	ProcessingCompletedAt *time.Time      `json:"processing_completed_at,omitempty" db:"processing_completed_at"`
	ProcessingError  string               `json:"processing_error,omitempty" db:"processing_error"`
	
	// Transcription
	HasTranscription bool                 `json:"has_transcription" db:"has_transcription"`
	TranscriptionID  *uuid.UUID           `json:"transcription_id,omitempty" db:"transcription_id"`
	
	// Quality metrics
	QualityScore     float64              `json:"quality_score,omitempty" db:"quality_score"`
	NoiseLevel       float64              `json:"noise_level,omitempty" db:"noise_level"`
	SilenceRatio     float64              `json:"silence_ratio,omitempty" db:"silence_ratio"`
	
	// Participants
	Participants     []Participant        `json:"participants,omitempty" db:"participants"`
	
	// Timestamps
	StartedAt        *time.Time           `json:"started_at,omitempty" db:"started_at"`
	EndedAt          *time.Time           `json:"ended_at,omitempty" db:"ended_at"`
	
	// Retention and compliance
	RetentionPeriod  time.Duration        `json:"retention_period" db:"retention_period"`
	ExpiresAt        *time.Time           `json:"expires_at,omitempty" db:"expires_at"`
	IsArchived       bool                 `json:"is_archived" db:"is_archived"`
	ArchivedAt       *time.Time           `json:"archived_at,omitempty" db:"archived_at"`
	
	// Tags and metadata
	Tags             []string             `json:"tags,omitempty" db:"tags"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	
	// Audit information
	CreatedBy        uuid.UUID            `json:"created_by" db:"created_by"`
	UpdatedBy        uuid.UUID            `json:"updated_by" db:"updated_by"`
	CreatedAt        time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at" db:"updated_at"`
}

// Participant represents a participant in a recording
type Participant struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Email    string    `json:"email,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	Role     string    `json:"role,omitempty"` // caller, callee, agent, customer
	Channel  int       `json:"channel,omitempty"`
	JoinedAt time.Time `json:"joined_at,omitempty"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
}

// RecordingSession represents a recording session
type RecordingSession struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	Description string                    `json:"description,omitempty" db:"description"`
	Status      RecordingStatus           `json:"status" db:"status"`
	
	// Session configuration
	AutoStart   bool                      `json:"auto_start" db:"auto_start"`
	AutoStop    bool                      `json:"auto_stop" db:"auto_stop"`
	MaxDuration time.Duration             `json:"max_duration" db:"max_duration"`
	
	// Audio settings
	Format      AudioFormat               `json:"format" db:"format"`
	SampleRate  int                       `json:"sample_rate" db:"sample_rate"`
	Channels    int                       `json:"channels" db:"channels"`
	BitDepth    int                       `json:"bit_depth" db:"bit_depth"`
	
	// Storage settings
	StorageProvider StorageProvider       `json:"storage_provider" db:"storage_provider"`
	StoragePath     string                `json:"storage_path" db:"storage_path"`
	
	// Processing settings
	EnableTranscription bool              `json:"enable_transcription" db:"enable_transcription"`
	EnableCompression   bool              `json:"enable_compression" db:"enable_compression"`
	EnableNoiseReduction bool             `json:"enable_noise_reduction" db:"enable_noise_reduction"`
	
	// Participants
	Participants    []Participant         `json:"participants,omitempty" db:"participants"`
	
	// Timestamps
	StartedAt       *time.Time            `json:"started_at,omitempty" db:"started_at"`
	EndedAt         *time.Time            `json:"ended_at,omitempty" db:"ended_at"`
	
	// Metadata
	Tags            []string              `json:"tags,omitempty" db:"tags"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	
	// Audit information
	CreatedBy       uuid.UUID             `json:"created_by" db:"created_by"`
	UpdatedBy       uuid.UUID             `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at" db:"updated_at"`
}

// Transcription represents a transcription of a recording
type Transcription struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	RecordingID uuid.UUID                 `json:"recording_id" db:"recording_id"`
	
	// Transcription content
	Text        string                    `json:"text" db:"text"`
	Segments    []TranscriptionSegment    `json:"segments,omitempty" db:"segments"`
	
	// Processing information
	Provider    string                    `json:"provider" db:"provider"`
	Language    string                    `json:"language" db:"language"`
	Confidence  float64                   `json:"confidence" db:"confidence"`
	
	// Speaker diarization
	HasDiarization bool                   `json:"has_diarization" db:"has_diarization"`
	SpeakerCount   int                    `json:"speaker_count" db:"speaker_count"`
	
	// Processing status
	Status      ProcessingStatus          `json:"status" db:"status"`
	StartedAt   *time.Time                `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time                `json:"completed_at,omitempty" db:"completed_at"`
	Error       string                    `json:"error,omitempty" db:"error"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// TranscriptionSegment represents a segment of transcription
type TranscriptionSegment struct {
	ID         int       `json:"id"`
	StartTime  float64   `json:"start_time"`
	EndTime    float64   `json:"end_time"`
	Text       string    `json:"text"`
	Speaker    string    `json:"speaker,omitempty"`
	Confidence float64   `json:"confidence"`
	Words      []Word    `json:"words,omitempty"`
}

// Word represents a word in transcription
type Word struct {
	Text       string  `json:"text"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Confidence float64 `json:"confidence"`
}

// RecordingMetadata represents metadata for a recording
type RecordingMetadata struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	RecordingID uuid.UUID                 `json:"recording_id" db:"recording_id"`
	
	// Metadata fields
	Key         string                    `json:"key" db:"key"`
	Value       string                    `json:"value" db:"value"`
	Type        string                    `json:"type" db:"type"` // string, number, boolean, json
	
	// Categorization
	Category    string                    `json:"category,omitempty" db:"category"`
	Tags        []string                  `json:"tags,omitempty" db:"tags"`
	
	// Audit information
	CreatedBy   uuid.UUID                 `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID                 `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// BatchJob represents a batch processing job
type BatchJob struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Type        string                    `json:"type" db:"type"` // transcribe, compress, migrate, cleanup
	Status      ProcessingStatus          `json:"status" db:"status"`
	
	// Job configuration
	Parameters  map[string]interface{}    `json:"parameters" db:"parameters"`
	RecordingIDs []uuid.UUID              `json:"recording_ids" db:"recording_ids"`
	
	// Progress tracking
	TotalItems     int                    `json:"total_items" db:"total_items"`
	ProcessedItems int                    `json:"processed_items" db:"processed_items"`
	FailedItems    int                    `json:"failed_items" db:"failed_items"`
	Progress       float64                `json:"progress" db:"progress"`
	
	// Timing
	StartedAt      *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	EstimatedCompletion *time.Time        `json:"estimated_completion,omitempty" db:"estimated_completion"`
	
	// Results
	Results        map[string]interface{} `json:"results,omitempty" db:"results"`
	Errors         []string               `json:"errors,omitempty" db:"errors"`
	
	// Audit information
	CreatedBy      uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// RecordingAudit represents an audit log entry for recordings
type RecordingAudit struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	RecordingID *uuid.UUID                `json:"recording_id,omitempty" db:"recording_id"`
	SessionID   *uuid.UUID                `json:"session_id,omitempty" db:"session_id"`
	
	// Audit information
	Action      string                    `json:"action" db:"action"`
	Actor       uuid.UUID                 `json:"actor" db:"actor"`
	ActorType   string                    `json:"actor_type" db:"actor_type"` // user, system, service
	
	// Request information
	IPAddress   string                    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   string                    `json:"user_agent,omitempty" db:"user_agent"`
	
	// Change information
	OldValues   map[string]interface{}    `json:"old_values,omitempty" db:"old_values"`
	NewValues   map[string]interface{}    `json:"new_values,omitempty" db:"new_values"`
	
	// Additional context
	Context     map[string]interface{}    `json:"context,omitempty" db:"context"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	
	// Timestamp
	Timestamp   time.Time                 `json:"timestamp" db:"timestamp"`
}

// StorageInfo represents storage information
type StorageInfo struct {
	Provider      StorageProvider `json:"provider"`
	TotalSpace    int64           `json:"total_space"`
	UsedSpace     int64           `json:"used_space"`
	FreeSpace     int64           `json:"free_space"`
	UsagePercent  float64         `json:"usage_percent"`
	FileCount     int64           `json:"file_count"`
	LastUpdated   time.Time       `json:"last_updated"`
}

// RecordingStatistics represents recording statistics
type RecordingStatistics struct {
	TotalRecordings    int64                    `json:"total_recordings"`
	ActiveRecordings   int64                    `json:"active_recordings"`
	CompletedRecordings int64                   `json:"completed_recordings"`
	FailedRecordings   int64                    `json:"failed_recordings"`
	
	// Storage statistics
	TotalStorageUsed   int64                    `json:"total_storage_used"`
	StorageByProvider  map[StorageProvider]int64 `json:"storage_by_provider"`
	
	// Format statistics
	RecordingsByFormat map[AudioFormat]int64    `json:"recordings_by_format"`
	
	// Time range
	StartTime          time.Time                `json:"start_time"`
	EndTime            time.Time                `json:"end_time"`
}

// CreateRecordingRequest represents a request to create a recording
type CreateRecordingRequest struct {
	SessionID   *uuid.UUID                `json:"session_id,omitempty"`
	CallID      *uuid.UUID                `json:"call_id,omitempty"`
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description,omitempty"`
	Type        RecordingType             `json:"type" binding:"required"`
	Format      AudioFormat               `json:"format,omitempty"`
	SampleRate  int                       `json:"sample_rate,omitempty"`
	Channels    int                       `json:"channels,omitempty"`
	BitDepth    int                       `json:"bit_depth,omitempty"`
	StorageProvider StorageProvider       `json:"storage_provider,omitempty"`
	Participants []Participant            `json:"participants,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// UpdateRecordingRequest represents a request to update a recording
type UpdateRecordingRequest struct {
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Status      *RecordingStatus          `json:"status,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// StartRecordingRequest represents a request to start recording
type StartRecordingRequest struct {
	SessionID   uuid.UUID                 `json:"session_id" binding:"required"`
	Format      AudioFormat               `json:"format,omitempty"`
	SampleRate  int                       `json:"sample_rate,omitempty"`
	Channels    int                       `json:"channels,omitempty"`
	BitDepth    int                       `json:"bit_depth,omitempty"`
	MaxDuration time.Duration             `json:"max_duration,omitempty"`
	AutoStop    bool                      `json:"auto_stop,omitempty"`
	EnableTranscription bool              `json:"enable_transcription,omitempty"`
	Participants []Participant            `json:"participants,omitempty"`
}

// StopRecordingRequest represents a request to stop recording
type StopRecordingRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
	Reason    string    `json:"reason,omitempty"`
}

// UploadChunkRequest represents a request to upload an audio chunk
type UploadChunkRequest struct {
	SessionID    uuid.UUID `json:"session_id" binding:"required"`
	ChunkIndex   int       `json:"chunk_index" binding:"required"`
	ChunkData    []byte    `json:"chunk_data" binding:"required"`
	IsLastChunk  bool      `json:"is_last_chunk"`
	Timestamp    time.Time `json:"timestamp"`
}

// CreateSessionRequest represents a request to create a recording session
type CreateSessionRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description,omitempty"`
	AutoStart   bool                      `json:"auto_start,omitempty"`
	AutoStop    bool                      `json:"auto_stop,omitempty"`
	MaxDuration time.Duration             `json:"max_duration,omitempty"`
	Format      AudioFormat               `json:"format,omitempty"`
	SampleRate  int                       `json:"sample_rate,omitempty"`
	Channels    int                       `json:"channels,omitempty"`
	BitDepth    int                       `json:"bit_depth,omitempty"`
	StorageProvider StorageProvider       `json:"storage_provider,omitempty"`
	EnableTranscription bool              `json:"enable_transcription,omitempty"`
	EnableCompression bool                `json:"enable_compression,omitempty"`
	EnableNoiseReduction bool             `json:"enable_noise_reduction,omitempty"`
	Participants []Participant            `json:"participants,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// UpdateSessionRequest represents a request to update a recording session
type UpdateSessionRequest struct {
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Status      *RecordingStatus          `json:"status,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// BatchProcessRequest represents a request for batch processing
type BatchProcessRequest struct {
	Type         string                    `json:"type" binding:"required"` // transcribe, compress, migrate, cleanup
	RecordingIDs []uuid.UUID               `json:"recording_ids" binding:"required"`
	Parameters   map[string]interface{}    `json:"parameters,omitempty"`
}

// SearchMetadataRequest represents a request to search metadata
type SearchMetadataRequest struct {
	Query       string                    `json:"query,omitempty"`
	Filters     map[string]interface{}    `json:"filters,omitempty"`
	Categories  []string                  `json:"categories,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	DateFrom    *time.Time                `json:"date_from,omitempty"`
	DateTo      *time.Time                `json:"date_to,omitempty"`
	Limit       int                       `json:"limit,omitempty"`
	Offset      int                       `json:"offset,omitempty"`
}

// ExportAuditRequest represents a request to export audit logs
type ExportAuditRequest struct {
	Format      string     `json:"format" binding:"required"` // json, csv, xml
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Actions     []string   `json:"actions,omitempty"`
	Actors      []uuid.UUID `json:"actors,omitempty"`
	RecordingIDs []uuid.UUID `json:"recording_ids,omitempty"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	SessionID uuid.UUID              `json:"session_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// RealtimeStatus represents real-time recording status
type RealtimeStatus struct {
	SessionID       uuid.UUID       `json:"session_id"`
	Status          RecordingStatus `json:"status"`
	Duration        time.Duration   `json:"duration"`
	FileSize        int64           `json:"file_size"`
	QualityScore    float64         `json:"quality_score,omitempty"`
	NoiseLevel      float64         `json:"noise_level,omitempty"`
	ParticipantCount int            `json:"participant_count"`
	LastUpdate      time.Time       `json:"last_update"`
}