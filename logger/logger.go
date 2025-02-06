package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger just wraps a zap.SugaredLogger.
type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() (Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return Logger{}, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return Logger{
		SugaredLogger: logger.Sugar(),
	}, nil
}
