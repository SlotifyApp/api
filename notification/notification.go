package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/logger"
)

// ClientSet represents a set of clients for a user.
type ClientSet map[http.ResponseWriter]struct{}

// Service shows behaviour for a notification service impl.
type Service interface {
	DeleteUserConn(logger *logger.Logger, userID uint32, w http.ResponseWriter)
	RegisterUserClient(logger *logger.Logger, userID uint32, w http.ResponseWriter) error
	SendNotification(logger *logger.Logger, db *database.Database,
		userID uint32, notif database.CreateNotificationParams) error
}

// SSENotificationService is a Server-Side Events notification service impl.
type SSENotificationService struct {
	// Maps a userID to a set of clients that can be used to send notifications to.
	conns map[uint32]ClientSet

	// Need a lock as many goroutines may be affecting these maps.
	mu sync.Mutex
}

// NewSSENotificationService creates a new instance of SSENotificationService.
func NewSSENotificationService() *SSENotificationService {
	return &SSENotificationService{
		conns: make(map[uint32]ClientSet),
	}
}

// RegisterUserClient registers a user client to send notifications to.
func (sse *SSENotificationService) RegisterUserClient(logger *logger.Logger,
	userID uint32, w http.ResponseWriter,
) error {
	if w == nil {
		return ErrNotifClientNil
	}

	sse.mu.Lock()

	defer sse.mu.Unlock()
	if sse.conns[userID] == nil {
		sse.conns[userID] = make(ClientSet)
	}

	clientSet := sse.conns[userID]

	// add client
	clientSet[w] = struct{}{}

	logger.Infof("Successfully added client for userID id(%d), clients: %+v", userID, clientSet)

	return nil
}

// DeleteUserClients attempts to deletes a user from the conns map, if there is no user this is a no-op.
func (sse *SSENotificationService) DeleteUserConn(logger *logger.Logger, userID uint32, w http.ResponseWriter) {
	sse.mu.Lock()

	defer sse.mu.Unlock()

	clientSet := sse.conns[userID]

	logger.Info("Deleting user id(%d) connection, clients: %+v", userID, clientSet)

	delete(sse.conns[userID], w)
}

// Store the notification in the database.
func storeNotification(ctx context.Context, db *database.Database,
	userID uint32, notif database.CreateNotificationParams,
) (*database.Notification, error) {
	notifID, err := db.CreateNotification(ctx, notif)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, fmt.Errorf("context cancelled during creating notif in notification table: %w", err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, fmt.Errorf("deadline exceeded during creating notif in notification table: %w", err)
		default:
			return nil, fmt.Errorf("failed to add notification to notification table: %w", err)
		}
	}

	// Add to user table
	dbParams := database.CreateUserNotificationParams{
		UserID: userID,
		//nolint: gosec // id is unsigned 32 bit int
		NotificationID: uint32(notifID),
	}
	rows, err := db.CreateUserNotification(ctx, dbParams)

	if rows != 1 {
		return nil, database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
	}

	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, fmt.Errorf("context cancelled during creating notif in UserToNotification table: %w", err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, fmt.Errorf("deadline exceeded during creating notif in UserToNotification table: %w", err)
		default:
			return nil, fmt.Errorf("failed to add notification to UserToNotification table: %w", err)
		}
	}

	return &database.Notification{
		//nolint: gosec // id is unsigned 32 bit int
		ID:      uint32(notifID),
		Message: notif.Message,
		Created: notif.Created,
	}, nil
}

// SendNotification sends a notification to ALL clients of a user.
// The notification is also stored in the database regardless of whether the user has a client or not.
func (sse *SSENotificationService) SendNotification(logger *logger.Logger,
	db *database.Database, userID uint32, notif database.CreateNotificationParams,
) error {
	sse.mu.Lock()
	clients := sse.conns[userID]

	sse.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
	defer cancel()

	storedNotif, err := storeNotification(ctx, db, userID, notif)
	if err != nil {
		return fmt.Errorf("failed to store notification for user: %w", err)
	}

	if clients == nil {
		// No clients available for user so store the notification
		logger.Infof("user id(%d) does not have a client", userID)
		return nil
	}

	log.Printf("Sending notification: clients: %+v, len(clients): %d", clients, len(clients))

	for c := range clients {
		logger.Info("attempting to flush to client")
		if c == nil {
			logger.Info("client was nil, attempting to delete")
			sse.DeleteUserConn(logger, userID, c)
			continue
		}
		var notifJSON []byte
		if notifJSON, err = json.Marshal(*storedNotif); err != nil {
			return fmt.Errorf("failed to encode notification as json: %w", err)
		}

		fmt.Fprintf(c, "event: calendar_notification\n")
		fmt.Fprintf(c, "data: %s\n\n", notifJSON)

		f, ok := c.(http.Flusher)

		if !ok {
			return errors.New("client doesn't implement flusher interface")
		}
		f.Flush()
	}
	return nil
}
