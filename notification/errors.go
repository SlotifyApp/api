package notification

import (
	"errors"
)

// ErrNotifClientNil is returned if the notification service attempts to register a nil client.
var ErrNotifClientNil = errors.New("notification client for user cannot be nil")
