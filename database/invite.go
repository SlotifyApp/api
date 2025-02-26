package database

import (
	"context"
	"errors"
	"fmt"
)

// CreateInviteWrapper is just a wrapper around db repository's CreateInvite to provide better error messages.
func CreateInviteWrapper(ctx context.Context, db *Database, params CreateInviteParams) error {
	rows, err := db.CreateInvite(ctx, params)
	if rows != 1 {
		return WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
	}
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("context cancelled creating invite: %w",
				err)
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("deadline exceeded during creating invite: %w", err)
		default:
			return fmt.Errorf("failed to create invite: %w", err)
		}
	}
	return nil
}

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
