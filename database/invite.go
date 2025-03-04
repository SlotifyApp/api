package database

import (
	"context"
	"errors"
	"fmt"
)

func DeleteInviteByIDWrapper(ctx context.Context, db *Database, inviteID uint32) error {
	rows, err := db.DeleteInviteByID(ctx, inviteID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("context cancelled deleting invite: %w",
				err)
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("deadline exceeded during deleting invite: %w", err)
		default:
			return fmt.Errorf("failed to delete invite: %w", err)
		}
	}

	if rows != 1 {
		return WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
	}

	return err
}
