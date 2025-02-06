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

func (MockNotificationService) DeleteUserConn(logger logger.Logger, userID uint32, w http.ResponseWriter) {
}

func (MockNotificationService) RegisterUserClient(logger *logger.Logger, userID uint32, w http.ResponseWriter) error {
	return nil
}

func (MockNotificationService) SendNotification(logger logger.Logger, db *database.Database,
	userID uint32, notif database.CreateNotificationParams,
) error {
	return nil
}

// ensure that we've conformed to the `NotificationService` with a compile-time check.
var _ notification.Service = (*MockNotificationService)(nil)
