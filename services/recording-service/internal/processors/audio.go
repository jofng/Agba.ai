package processors

import (
    "context"
    "go.uber.org/zap"
)

// AudioProcessor handles audio processing
type AudioProcessor struct {
    logger *zap.Logger
}

// NewAudioProcessor creates a new audio processor
func NewAudioProcessor(logger *zap.Logger) *AudioProcessor {
    return &AudioProcessor{logger: logger}
}

// Process processes audio data
func (p *AudioProcessor) Process(ctx context.Context, data []byte) ([]byte, error) {
    p.logger.Info("Processing audio data")
    return data, nil
}
