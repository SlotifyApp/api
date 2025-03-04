package cron_test

import (
	"context"
	"testing"
	"time"

	"github.com/SlotifyApp/slotify-backend/cron"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/stretchr/testify/require"
)

func TestRemoveWeekOldInvites(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute*5)
	defer cancel()

	l := testutil.NewLogger(t)
	db := testutil.NewDB(t, ctx)

	sg := testutil.InsertSlotifyGroup(t, db.DB)
	fromUser := testutil.InsertUser(t, db.DB)
	toUser := testutil.InsertUser(t, db.DB)
	// Create an invite that has expiration date in a month but is a week-old
	p := database.CreateInviteParams{
		SlotifyGroupID: sg.Id,
		FromUserID:     fromUser.Id,
		ToUserID:       toUser.Id,
		Message:        "Invite message blah",
		ExpiryDate:     time.Now().AddDate(0, 1, 0),
		CreatedAt:      time.Now().AddDate(0, 0, -8),
	}

	inviteCount := 100

	for range inviteCount {
		var err error
		_, err = db.CreateInvite(ctx, p)
		require.NoError(t, err, "createinvite test setup failed")
	}
	oldCount := testutil.GetWeekOldInviteCount(t, ctx, db)
	require.Equal(t, inviteCount, oldCount, "the correct number of invites were created during set up")

	cron.RemoveWeekOldInvites(ctx, db, l)

	newCount := testutil.GetWeekOldInviteCount(t, ctx, db)
	require.Equal(t, 0, newCount, "all week old invites were deleted")
}
