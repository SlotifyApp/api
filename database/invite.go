package database

import (
	"context"
	"errors"
	"fmt"
)

// CreateInvite is just a wrapper around db repository's CreateInvite to provide better error messages.
func CreateInvite(ctx context.Context, db *Database, params CreateInviteParams) error {
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
