package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

// Repository handles data operations for recording service
type Repository struct {
    db     *sql.DB
    logger *zap.Logger
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
    return &Repository{
        db:     db,
        logger: logger,
    }
}

// Recording represents a recording record
type Recording struct {
    ID        uuid.UUID              `json:"id"`
    SessionID string                 `json:"session_id"`
    UserID    *uuid.UUID             `json:"user_id,omitempty"`
    Filename  string                 `json:"filename"`
    FilePath  string                 `json:"file_path"`
    FileSize  int64                  `json:"file_size"`
    Duration  int                    `json:"duration"`
    Format    string                 `json:"format"`
    Quality   string                 `json:"quality"`
    Status    string                 `json:"status"`
    Metadata  map[string]interface{} `json:"metadata"`
    StartedAt time.Time              `json:"started_at"`
    EndedAt   *time.Time             `json:"ended_at,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}

// CreateRecording creates a new recording
func (r *Repository) CreateRecording(ctx context.Context, recording *Recording) error {
    query := `
        INSERT INTO recordings (id, session_id, user_id, filename, file_path, file_size, duration, format, quality, status, metadata, started_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

    metadataJSON, err := json.Marshal(recording.Metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    now := time.Now()
    recording.CreatedAt = now
    recording.UpdatedAt = now

    _, err = r.db.ExecContext(ctx, query,
        recording.ID, recording.SessionID, recording.UserID,
        recording.Filename, recording.FilePath, recording.FileSize,
        recording.Duration, recording.Format, recording.Quality,
        recording.Status, metadataJSON, recording.StartedAt, now, now)

    if err != nil {
        r.logger.Error("Failed to create recording", zap.Error(err))
        return fmt.Errorf("failed to create recording: %w", err)
    }

    return nil
}

// GetRecording retrieves a recording by ID
func (r *Repository) GetRecording(ctx context.Context, id uuid.UUID) (*Recording, error) {
    query := `
        SELECT id, session_id, user_id, filename, file_path, file_size, 
               duration, format, quality, status, metadata, started_at, 
               ended_at, created_at, updated_at
        FROM recordings WHERE id = $1`

    var recording Recording
    var metadataJSON []byte
    var userID sql.NullString
    var endedAt sql.NullTime

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &recording.ID, &recording.SessionID, &userID,
        &recording.Filename, &recording.FilePath, &recording.FileSize,
        &recording.Duration, &recording.Format, &recording.Quality,
        &recording.Status, &metadataJSON, &recording.StartedAt,
        &endedAt, &recording.CreatedAt, &recording.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("recording not found: %s", id.String())
        }
        return nil, fmt.Errorf("failed to get recording: %w", err)
    }

    // Handle nullable fields
    if userID.Valid {
        uid, _ := uuid.Parse(userID.String)
        recording.UserID = &uid
    }
    if endedAt.Valid {
        recording.EndedAt = &endedAt.Time
    }

    // Unmarshal metadata
    if err := json.Unmarshal(metadataJSON, &recording.Metadata); err != nil {
        r.logger.Warn("Failed to unmarshal metadata", zap.Error(err))
        recording.Metadata = make(map[string]interface{})
    }

    return &recording, nil
}

// UpdateRecordingStatus updates the status of a recording
func (r *Repository) UpdateRecordingStatus(ctx context.Context, id uuid.UUID, status string) error {
    query := `
        UPDATE recordings 
        SET status = $2, updated_at = $3
        WHERE id = $1`

    _, err := r.db.ExecContext(ctx, query, id, status, time.Now())
    if err != nil {
        r.logger.Error("Failed to update recording status", zap.Error(err))
        return fmt.Errorf("failed to update recording status: %w", err)
    }

    return nil
}

// HealthCheck checks repository health
func (r *Repository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
