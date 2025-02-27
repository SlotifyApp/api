package database

import (
	"context"
	"errors"
	"fmt"
)

func CheckMemberInSlotifyGroupWrapper(ctx context.Context,
	db *Database, arg CheckMemberInSlotifyGroupParams,
) (bool, error) {
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

func AddUserToSlotifyGroupWrapper(ctx context.Context, qtx *Queries, arg AddUserToSlotifyGroupParams) error {
	rowsAffected, err := qtx.AddUserToSlotifyGroup(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	if rowsAffected != 1 {
		err = WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{1},
		}

		return fmt.Errorf("failed to add user to slotifyGroup: %w", err)
	}

	return nil
}
