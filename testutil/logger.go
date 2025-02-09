package testutil

import (
	"testing"

	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/stretchr/testify/require"
)

// NewLogger creates a Logger instance, fails if there is an error.
func NewLogger(t *testing.T) *logger.Logger {
	l, err := logger.NewLogger()
	require.NoError(t, err, "new logger error")

	return l
}
