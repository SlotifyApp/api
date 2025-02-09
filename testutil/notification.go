package testutil

import (
	"fmt"
	"time"
)

// GetExpectedNotificationSSE returns the expected notification that was flushed to the client.
func GetExpectedNotificationSSE(notifID int64, message string, created time.Time) string {
	return fmt.Sprintf(
		"event: calendar_notification\ndata: {\"id\":%d,\"message\":\"%s\",\"created\":\"%s\"}\n\n",
		notifID, message, created.Format(time.RFC3339Nano))
}
