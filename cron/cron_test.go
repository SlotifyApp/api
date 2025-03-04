package cron_test

import (
	"context"
	"testing"
	"time"

	"github.com/SlotifyApp/slotify-backend/cron"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestRemoveWeekOldInvites(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute*5)
	defer cancel()

	l := testutil.NewLogger(t)
	db := testutil.NewDB(t, ctx)

	sg := testutil.InsertSlotifyGroup(t, db.DB)
	fromUser := testutil.InsertUser(t, db.DB)
	toUser := testutil.InsertUser(t, db.DB)
	// Create an invite that has expiration date in a month but is a week-old
	weekOldP := database.CreateInviteParams{
		SlotifyGroupID: sg.Id,
		FromUserID:     fromUser.Id,
		ToUserID:       toUser.Id,
		Message:        "Invite message blah",
		ExpiryDate:     time.Now().AddDate(0, 1, 0),
		Status:         database.InviteStatusPending,
		CreatedAt:      time.Now().AddDate(0, 0, -8),
	}

	inviteCount := 100

	// Setup
	for range inviteCount - 1 {
		var err error
		_, err = db.CreateInvite(ctx, weekOldP)
		require.NoError(t, err, "createinvite test setup failed")
	}

	moreThanWeekOldP := database.CreateInviteParams{
		SlotifyGroupID: sg.Id,
		FromUserID:     fromUser.Id,
		ToUserID:       toUser.Id,
		Message:        "Invite message blah",
		Status:         database.InviteStatusPending,
		ExpiryDate:     time.Now().AddDate(0, 1, 0),
		CreatedAt:      time.Now().AddDate(0, -1, 0),
	}

	var err error
	_, err = db.CreateInvite(ctx, moreThanWeekOldP)
	require.NoError(t, err, "createinvite test setup failed")

	// Assert
	oldCount := testutil.GetWeekOldInviteCount(t, ctx, db)
	require.Equal(t, inviteCount, oldCount, "the correct number of invites were created during set up")

	cron.RemoveWeekOldInvites(ctx, db, l)

	newCount := testutil.GetWeekOldInviteCount(t, ctx, db)
	require.Equal(t, 0, newCount, "all week old invites were deleted")
}

func TestRemoveWeekOldNotifications(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute*5)
	defer cancel()

	l := testutil.NewLogger(t)
	db := testutil.NewDB(t, ctx)

	u := testutil.InsertUser(t, db.DB)

	notifCount := 100

	// Setup
	for range notifCount - 1 {
		p := database.CreateNotificationParams{
			Message: gofakeit.MovieName(),
			Created: time.Now().AddDate(0, 0, -8),
		}
		testutil.CreateNotificationAndLink(t, ctx, db, p, u.Id)
	}

	p := database.CreateNotificationParams{
		Message: gofakeit.MovieName(),
		Created: time.Now().AddDate(0, -1, 0),
	}
	testutil.CreateNotificationAndLink(t, ctx, db, p, u.Id)

	// Assert
	oldCount := testutil.GetWeekOldNotificationCount(t, ctx, db)
	require.Equal(t, notifCount, oldCount, "the correct number of notifications were created during set up")

	cron.RemoveWeekOldNotifications(ctx, db, l)

	newCount := testutil.GetWeekOldNotificationCount(t, ctx, db)
	require.Equal(t, 0, newCount, "all week old notifications were deleted")
}

func TestExpireInvites(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute*5)
	defer cancel()

	l := testutil.NewLogger(t)
	db := testutil.NewDB(t, ctx)

	sg := testutil.InsertSlotifyGroup(t, db.DB)
	fromUser := testutil.InsertUser(t, db.DB)
	toUser := testutil.InsertUser(t, db.DB)

	expiredInviteCount := 7
	// Setup
	for i := 1; i <= expiredInviteCount; i++ {
		p := database.CreateInviteParams{
			SlotifyGroupID: sg.Id,
			FromUserID:     fromUser.Id,
			ToUserID:       toUser.Id,
			Message:        "Invite message blah",
			ExpiryDate:     time.Now().AddDate(0, 0, -1*i),
			Status:         database.InviteStatusPending,
			CreatedAt:      time.Now().AddDate(0, 0, 0),
		}
		var err error
		_, err = db.CreateInvite(ctx, p)
		require.NoError(t, err, "createinvite test setup failed")
	}

	// Assert
	oldCount := testutil.GetExpiredInvitesCount(t, ctx, db)

	cron.ExpireInvites(ctx, db, l)

	newCount := testutil.GetExpiredInvitesCount(t, ctx, db)

	require.Equal(t, oldCount+expiredInviteCount, newCount, "all relevant invites were set to expired")
}
