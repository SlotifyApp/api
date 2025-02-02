package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/brianvoe/gofakeit/v7"
	openapi_types "github.com/oapi-codegen/runtime/types"
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
	require.NoError(t, err, "unable to form query")
	defer func() {
		err = rows.Close()
		require.NoError(t, err, "unable to close rows")
	}()

	var users api.Users
	for rows.Next() {
		var uq api.UserQuery
		err = rows.Scan(&uq.Id, &uq.Email, &uq.FirstName, &uq.LastName, &uq.HomeAccountID)
		require.NoError(t, err, "unable to scan rows")
		users = append(users, uq.User)
	}

	err = rows.Err()
	require.NoError(t, err, "sql rows error")

	return users
}

func GetTeamRows(t *testing.T, db *sql.DB) api.Teams {
	rows, err := db.Query("SELECT * FROM Team")
	require.NoError(t, err, "unable to form query")
	defer func() {
		err = rows.Close()
		require.NoError(t, err, "unable to close rows")
	}()

	var teams api.Teams
	for rows.Next() {
		var team api.Team
		err = rows.Scan(&team.Id, &team.Name)
		require.NoError(t, err, "unable to scan rows")
		teams = append(teams, team)
	}

	err = rows.Err()
	require.NoError(t, err, "sql rows error")

	return teams
}

// GetNextAutoIncrementValue gets the next auto increment value for a table,
// this is different to just getting count.
func GetNextAutoIncrementValue(t *testing.T, db *sql.DB, tableName string) int {
	dbName, present := os.LookupEnv("DB_NAME")
	require.True(t, present, "DB_NAME is in env variables")
	query := `
		SELECT AUTO_INCREMENT
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_name = ?;
	`
	var nextAutoIncrement int
	err := db.QueryRow(query, dbName, tableName).Scan(&nextAutoIncrement)
	require.NoError(t, err, "scanning row fails")
	return nextAutoIncrement
}

func AddUserToTeam(t *testing.T, db *sql.DB, userID int, teamID int) {
	res, err := db.Exec("INSERT INTO UserToTeam (user_id, team_id) VALUES (?, ?)", userID, teamID)
	require.NoError(t, err, "failed to execute sql query to add user to team")

	rows, err := res.RowsAffected()
	require.NoError(t, err, "failed to get the number of rows affected")

	require.Equal(t, int64(1), rows, "rows returned is not correct")
}

func InsertTeam(t *testing.T, db *sql.DB) api.Team {
	name := gofakeit.ProductName()

	res, err := db.Exec("INSERT INTO Team (name) VALUES (?)", name)
	require.NoError(t, err, "db insert team failed")

	rows, err := res.RowsAffected()
	require.Equal(t, int64(1), rows, "rows affected after insert is 1")
	require.NoError(t, err, "failed to get rows affected")

	id, err := res.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return api.Team{
		Id:   int(id),
		Name: name,
	}
}

// NewServerAndDB creates a server and a db, test fails
// if any errors are returned.
func NewServerAndDB(t *testing.T, ctx context.Context) (*sql.DB, *api.Server) {
	db, err := database.NewDatabaseWithContext(ctx)
	require.NoError(t, err, "error creating database handle")
	require.NotNil(t, db, "db handle cannot be nil")

	server, err := api.NewServerWithContext(ctx, db, api.WithNotInitMSALClient())

	require.NoError(t, err, "error creating server ")
	require.NotNil(t, db, "server cannot be nil")

	return db, server
}

// CloseDB closes the given DB handle, often used with 'defer'.
func CloseDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("testutil: failed to close db: %s", err.Error())
	}
}

type userOptions struct {
	firstName *string
	lastName  *string
	email     *openapi_types.Email
}

type UserOption func(opts *userOptions) error

func WithFirstName(firstName string) UserOption {
	return func(options *userOptions) error {
		options.firstName = &firstName
		return nil
	}
}

func WithLastName(lastName string) UserOption {
	return func(options *userOptions) error {
		options.lastName = &lastName
		return nil
	}
}

func WithEmail(email openapi_types.Email) UserOption {
	return func(options *userOptions) error {
		options.email = &email
		return nil
	}
}

func InsertUser(t *testing.T, db *sql.DB, userOpts ...UserOption) api.User {
	var opts userOptions
	for _, opt := range userOpts {
		err := opt(&opts)
		require.NoError(t, err, "failed to apply user option when creating user")
	}
	var firstName string
	if opts.firstName == nil {
		firstName = gofakeit.FirstName()
	} else {
		firstName = *opts.firstName
	}

	var lastName string
	if opts.lastName == nil {
		lastName = gofakeit.LastName()
	} else {
		lastName = *opts.lastName
	}

	var email string
	if opts.email == nil {
		email = gofakeit.Email()
	} else {
		email = string(*opts.email)
	}

	res, err := db.Exec("INSERT INTO User (email, first_name, last_name) VALUES (?, ?, ?)", email, firstName, lastName)
	require.NoError(t, err, "db insert user failed")

	rows, err := res.RowsAffected()
	require.Equal(t, int64(1), rows, "rows affected after insert is 1")
	require.NoError(t, err, "failed to get rows affected")

	id, err := res.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return api.User{
		Id:        int(id),
		Email:     openapi_types.Email(email),
		FirstName: firstName,
		LastName:  lastName,
	}
}
