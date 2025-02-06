package testutil

import (
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

func (MockNotificationService) SendNotification(_ *logger.Logger, _ *database.Database,
	_ uint32, _ database.CreateNotificationParams,
) error {
	return nil
}

// ensure that we've conformed to the `NotificationService` with a compile-time check.
var _ notification.Service = (*MockNotificationService)(nil)
