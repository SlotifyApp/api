package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/mocks"
	"github.com/SlotifyApp/slotify-backend/notification"
	"github.com/brianvoe/gofakeit/v7"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

func AddUserToTeam(t *testing.T, db *sql.DB, userID uint32, teamID uint32) {
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
		//nolint: gosec // id is unsigned 32 bit int
		Id:   uint32(id),
		Name: name,
	}
}

type options struct {
	notificationService notification.Service
}

func WithNotificationService(notifService notification.Service) TestServerOption {
	return func(options *options) error {
		if notifService == nil {
			return errors.New("notifService must not be nil")
		}
		options.notificationService = notifService
		return nil
	}
}

type TestServerOption func(opts *options) error

// NewServerAndDB creates a server and a db, test fails
// if any errors are returned.
func NewServerAndDB(t *testing.T,
	ctx context.Context,
	testServerOpts ...TestServerOption,
) (*database.Database, *api.Server) {
	var opts options
	for _, opt := range testServerOpts {
		err := opt(&opts)
		require.NoError(t, err, "failure applying options")
	}

	var notificationService notification.Service
	if opts.notificationService == nil {
		ctrl := gomock.NewController(t)
		notificationService = mocks.NewMockService(ctrl)
	} else {
		notificationService = opts.notificationService
	}

	db, err := database.NewDatabaseWithContext(ctx)
	require.NoError(t, err, "error creating database handle")
	require.NotNil(t, db, "db handle cannot be nil")

	server, err := api.NewServerWithContext(ctx, db,
		api.WithNotInitMSALClient(), api.WithNotificationService(notificationService))

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
		//nolint: gosec // id is unsigned 32 bit int
		Id:        uint32(id),
		Email:     openapi_types.Email(email),
		FirstName: firstName,
		LastName:  lastName,
	}
}
