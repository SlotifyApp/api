package testutil

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SlotifyApp/slotify-backend/api"
)

// GetCount gets the row count of a given SQL table.
func GetCount(dbh *sql.DB, table string) (int, error) {
	var count int
	//nolint: gosec //This is a test helper, not used in actual production.
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := dbh.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GetUserRows returns all rows in the User table.
func GetUserRows(dbh *sql.DB) (api.Users, error) {
	query := "SELECT * FROM User"
	rows, err := dbh.Query(query)
	if err != nil {
		return api.Users{}, nil
	}
	defer func() {
		if err = rows.Close(); err != nil {
			log.Printf("failed to close sql rows: %s", err.Error())
		}
	}()

	var users api.Users
	for rows.Next() {
		var user api.User
		if err = rows.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName); err != nil {
			return api.Users{}, nil
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return api.Users{}, nil
	}
	return users, nil
}
