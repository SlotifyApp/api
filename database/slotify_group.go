package database

import (
	"context"
	"errors"
	"fmt"
)

func CheckMemberIsInSlotifyGroup(ctx context.Context, db *Database, arg CheckMemberInSlotifyGroupParams) (bool, error) {
	// Check that the user creating the event is in the group
	isUserInGroup, err := db.CheckMemberInSlotifyGroup(ctx, CheckMemberInSlotifyGroupParams{
		UserID:         arg.UserID,
		SlotifyGroupID: arg.SlotifyGroupID,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return false, fmt.Errorf("context cancelled, checking member in slotify group: %w",
				err)
		case errors.Is(err, context.DeadlineExceeded):
			return false, fmt.Errorf("deadline exceeded, checking member in slotify group: %w", err)
		default:
			return false, fmt.Errorf("failed checking member in slotify group: %w", err)
		}
	}

	return isUserInGroup > 0, nil
}
