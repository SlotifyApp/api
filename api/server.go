package api

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/SlotifyApp/slotify-backend/notification"
)

type options struct {
	logger              *logger.Logger
	msalClient          *confidential.Client
	initMSALClient      *bool
	notificationService notification.Service
}

type ServerOption func(opts *options) error

func WithLogger(logger *logger.Logger) ServerOption {
	return func(options *options) error {
		if logger == nil {
			return errors.New("logger must not be nil")
		}
		options.logger = logger
		return nil
	}
}

func WithNotificationService(notifService notification.Service) ServerOption {
	return func(options *options) error {
		if notifService == nil {
			return errors.New("notifService must not be nil")
		}
		options.notificationService = notifService
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
	Logger              *logger.Logger
	DB                  *database.Database
	MSALClient          *confidential.Client
	NotificationService notification.Service
}

func NewServerWithContext(_ context.Context, db *database.Database, serverOpts ...ServerOption) (*Server, error) {
	var opts options
	for _, opt := range serverOpts {
		err := opt(&opts)
		if err != nil {
			return nil, err
		}
	}
	var serverLogger logger.Logger
	if opts.logger == nil {
		var err error
		if serverLogger, err = logger.NewLogger(); err != nil {
			log.Fatalf("failed to create new logger: %s", err.Error())
		}
	} else {
		serverLogger = *opts.logger
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

	var notificationService notification.Service
	if opts.notificationService == nil {
		notificationService = notification.NewSSENotificationService()
	} else {
		notificationService = opts.notificationService
	}

	if db == nil {
		return nil, errors.New("db must be provided")
	}

	return &Server{
		Logger:              &serverLogger,
		DB:                  db,
		MSALClient:          msalClient,
		NotificationService: notificationService,
	}, nil
}
