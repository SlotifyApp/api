package api

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/SlotifyApp/slotify-backend/database"
)

type options struct {
	logger         *zap.SugaredLogger
	msalClient     *confidential.Client
	initMSALClient *bool
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

func WithMSALClient(msalClient *confidential.Client) ServerOption {
	return func(options *options) error {
		if msalClient == nil {
			return errors.New("msalClient must not be nil")
		}
		options.msalClient = msalClient
		return nil
	}
}

// WithNotInitMSALClient prevents setting MSAL client if it has not been passed in.
func WithNotInitMSALClient() ServerOption {
	return func(options *options) error {
		shouldInit := false
		options.initMSALClient = &shouldInit
		return nil
	}
}

type Server struct {
	Logger     *zap.SugaredLogger
	DB         *database.Database
	MSALClient *confidential.Client
}

func NewServerWithContext(_ context.Context, db *database.Database, serverOpts ...ServerOption) (*Server, error) {
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

	// default to initialising MSAL client unless specified
	var shouldInitMSALClient bool
	if opts.initMSALClient == nil {
		shouldInitMSALClient = true
	} else {
		shouldInitMSALClient = *opts.initMSALClient
	}

	var msalClient *confidential.Client
	if opts.msalClient == nil && shouldInitMSALClient {
		// Create new msal client
		c, err := createMSALClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create msal client: %w", err)
		}
		msalClient = &c
	} else {
		msalClient = opts.msalClient
	}

	if db == nil {
		return nil, errors.New("db must be provided")
	}

	return &Server{
		Logger:     serverLogger,
		DB:         db,
		MSALClient: msalClient,
	}, nil
}
