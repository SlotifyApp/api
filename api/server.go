package api

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"go.uber.org/zap"
)

type options struct {
	logger *zap.SugaredLogger
}

type ServerOption func(opts *options) error

func WithLogger(logger *zap.SugaredLogger) ServerOption {
	return func(options *options) error {
		if logger == nil {
			return errors.New("logger must not be nil")
		}
		options.logger = logger
		return nil
	}
}

type Server struct {
	Logger                           *zap.SugaredLogger
	DB                               *sql.DB
	RefreshTokenRepository           RefreshTokenRepository
	TeamRepository                   TeamRepository
	UserRepository                   UserRepository
	UserToMSFTRefreshTokenRepository UserToMSFTRefreshTokenRepository
}

func NewServerWithContext(_ context.Context, db *sql.DB, serverOpts ...ServerOption) (*Server, error) {
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
		defer func() {
			if err := logger.Sync(); err != nil {
				log.Printf("failed to sync zap logger: %s", err.Error())
			}
		}()
		serverLogger = logger.Sugar()
	} else {
		serverLogger = opts.logger
	}

	if db == nil {
		return nil, errors.New("db must be provided")
	}

	refreshTokenRepository := NewRefreshTokenRepository(serverLogger, db)
	teamRepository := NewTeamRepository(serverLogger, db)
	userRepository := NewUserRepository(serverLogger, db)
	userToMSFTRefreshTokenRepository := NewUserToMSFTRefreshTokenRepository(serverLogger, db)

	return &Server{
		Logger:                           serverLogger,
		DB:                               db,
		RefreshTokenRepository:           refreshTokenRepository,
		TeamRepository:                   teamRepository,
		UserRepository:                   userRepository,
		UserToMSFTRefreshTokenRepository: userToMSFTRefreshTokenRepository,
	}, nil
}
