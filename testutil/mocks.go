package testutil

import (
	"context"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/SlotifyApp/slotify-backend/notification"
)

// MockNotificationService implements the notification.NotificationService interface.
// This is used for testing.
type MockNotificationService struct{}

func (MockNotificationService) DeleteUserConn(_ *logger.Logger, _ uint32, _ http.ResponseWriter) {
}

func (MockNotificationService) RegisterUserClient(_ *logger.Logger, _ uint32, _ http.ResponseWriter) error {
	return nil
}

func (MockNotificationService) SendNotification(_ context.Context, _ *logger.Logger, _ database.NotificationDatabase,
	_ uint32, _ database.CreateNotificationParams,
) error {
	return nil
}

// ensure that we've conformed to the `NotificationService` with a compile-time check.
var _ notification.Service = (*MockNotificationService)(nil)

// MockNotificationDatabase implements the database.NotificationDatabase interface.
type MockNotificationDatabase struct {
	NotificationID int64
	AffectedRows   int64
}

func NewMockNotificationDatabase(affectedRows, notificationID int64) MockNotificationDatabase {
	return MockNotificationDatabase{
		AffectedRows:   affectedRows,
		NotificationID: notificationID,
	}
}

func (mnd MockNotificationDatabase) CreateNotification(_ context.Context,
	_ database.CreateNotificationParams,
) (int64, error) {
	return mnd.NotificationID, nil
}

func (mnd MockNotificationDatabase) CreateUserNotification(_ context.Context,
	_ database.CreateUserNotificationParams,
) (int64, error) {
	// if 1 isn't the number of affected rows
	return mnd.AffectedRows, nil
}

// ensure `MockNotificationDatabase` conforms to the `NotificationDatabase` interface with a compile-time check.
var _ database.NotificationDatabase = (*MockNotificationDatabase)(nil)
