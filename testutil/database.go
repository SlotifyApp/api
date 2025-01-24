package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/stretchr/testify/require"
)

// GetCount gets the row count of a given SQL table.
func GetCount(t *testing.T, db *sql.DB, table string) int {
	//nolint: gosec //This is a test helper, not used in actual production.
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)

	var count int
	err := db.QueryRow(query).Scan(&count)
	require.NoError(t, err, "unable to query row")

	return count
}

// GetUserRows returns all rows in the User table.
func GetUserRows(t *testing.T, db *sql.DB) api.Users {
	rows, err := db.Query("SELECT * FROM User")
	require.NoError(t, err, "unable to form query: %s", err.Error())
	defer func() {
		err = rows.Close()
		require.NoError(t, err, "unable to close rows")
	}()

	var users api.Users
	for rows.Next() {
		var user api.User
		err = rows.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName)
		require.NoError(t, err, "unable to scan rows")
		users = append(users, user)
	}

	err = rows.Err()
	require.NoError(t, err, "sql rows error")

	return users
}
