package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger creates a logger that wraps a zap.SugaredLogger.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger returns a new instance of Logger.
func NewLogger() (*Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}, nil
}
