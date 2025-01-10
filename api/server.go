package api

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type options struct {
	logger *zap.Logger
}

type ServerOption func(opts *options) error

func WithLogger(logger *zap.Logger) ServerOption {
	return func(options *options) error {
		if logger == nil {
			return errors.New("logger must not be nil")
		}
		options.logger = logger
		return nil
	}
}

type Server struct {
	Logger *zap.SugaredLogger
}

func NewServerWithContext(_ context.Context, serverOpts ...ServerOption) (*Server, error) {
	var opts options
	for _, opt := range serverOpts {
		err := opt(&opts)
		if err != nil {
			return nil, err
		}
	}
	var serverLogger *zap.SugaredLogger
	if opts.logger == nil {
		logger, _ := zap.NewProduction()
		//nolint:errcheck // This is taken from zap's docs
		defer logger.Sync()
		serverLogger = logger.Sugar()
	}

	return &Server{
		Logger: serverLogger,
	}, nil
}
