package testutil

import (
	"testing"

	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/stretchr/testify/require"
)

func NewLogger(t *testing.T) *logger.Logger {
	l, err := logger.NewLogger()
	require.NoError(t, err, "new logger error")

	return l
}
