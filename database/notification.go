package database

import (
	"context"
	"errors"
	"fmt"
)

// NotificationDatabase is an interface describing behaviour needed to correctly store a notification.
type NotificationDatabase interface {
	// CreateNotification returns the notification ID and an error.
	CreateNotification(ctx context.Context, arg CreateNotificationParams) (int64, error)
	// CreateUserNotification returns the number of affected rows and an error.
	CreateUserNotification(ctx context.Context, arg CreateUserNotificationParams) (int64, error)
}

// StoreNotification creates a notification in the 'Notification' table, and
// links it to a user via the 'UserToNotification' table.
func StoreNotification(ctx context.Context, db NotificationDatabase,
	userID uint32, notif CreateNotificationParams,
) (*Notification, error) {
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
	dbParams := CreateUserNotificationParams{
		UserID: userID,
		//nolint: gosec // id is unsigned 32 bit int
		NotificationID: uint32(notifID),
	}

	rows, err := db.CreateUserNotification(ctx, dbParams)

	if rows != 1 {
		return nil, WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
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

	return &Notification{
		//nolint: gosec // id is unsigned 32 bit int
		ID:      uint32(notifID),
		Message: notif.Message,
		Created: notif.Created,
	}, nil
}
