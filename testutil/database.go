// Package testutil implements common testing functionality.
//
// testutil functions take in a testing.T and this is used to fail the test if there is
// any error.
package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/mocks"
	"github.com/SlotifyApp/slotify-backend/notification"
	"github.com/avast/retry-go"
	"github.com/brianvoe/gofakeit/v7"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func CreateNotificationAndLink(t *testing.T, ctx context.Context, db *database.Database,
	arg database.CreateNotificationParams, userID uint32,
) {
	notificationID, err := db.CreateNotification(ctx, arg)

	require.NoError(t, err, "failed to create notification")

	_, err = db.CreateUserNotification(ctx,
		database.CreateUserNotificationParams{
			UserID: userID,
			//nolint: gosec // id is unsigned 32 bit int
			NotificationID: uint32(notificationID),
		})

	require.NoError(t, err, "failed to link user to notification")
}

// GetExpiredInvitesCount is a test wrapper for CountWeekOldInvites.
func GetExpiredInvitesCount(t *testing.T, ctx context.Context, db *database.Database) int {
	count, err := db.CountExpiredInvites(ctx)
	require.NoError(t, err, "unable to query row")

	return int(count)
}

// GetWeekOldNotificationCount is a test wrapper for CountWeekOldInvites.
func GetWeekOldNotificationCount(t *testing.T, ctx context.Context, db *database.Database) int {
	count, err := db.CountWeekOldNotifications(ctx)
	require.NoError(t, err, "unable to query row")

	return int(count)
}

// GetWeekOldInviteCount is a test wrapper for CountWeekOldInvites.
func GetWeekOldInviteCount(t *testing.T, ctx context.Context, db *database.Database) int {
	count, err := db.CountWeekOldInvites(ctx)
	require.NoError(t, err, "unable to query row")

	return int(count)
}

// GetCount gets the row count of a given SQL table.
func GetCount(t *testing.T, db *sql.DB, table string) int {
	//nolint: gosec //This is a test helper, not used in actual production.
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)

	var count int
	err := db.QueryRow(query).Scan(&count)
	require.NoError(t, err, "unable to query row")

	return count
}

// GetSlotifyGroupRows.
func GetSlotifyGroupRows(t *testing.T, db *sql.DB) api.SlotifyGroups {
	rows, err := db.Query("SELECT * FROM SlotifyGroup")
	require.NoError(t, err, "unable to form query")
	defer func() {
		err = rows.Close()
		require.NoError(t, err, "unable to close rows")
	}()

	var slotifyGroups api.SlotifyGroups
	for rows.Next() {
		var slotifyGroup api.SlotifyGroup
		err = rows.Scan(&slotifyGroup.Id, &slotifyGroup.Name)
		require.NoError(t, err, "unable to scan rows")
		slotifyGroups = append(slotifyGroups, slotifyGroup)
	}

	err = rows.Err()
	require.NoError(t, err, "sql rows error")

	return slotifyGroups
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

func AddUserToSlotifyGroup(t *testing.T, db *sql.DB, userID uint32, slotifyGroupID uint32) {
	res, err := db.Exec("INSERT INTO UserToSlotifyGroup (user_id, slotify_group_id) VALUES (?, ?)", userID, slotifyGroupID)
	require.NoError(t, err, "failed to execute sql query to add user to slotifyGroup")

	rows, err := res.RowsAffected()
	require.NoError(t, err, "failed to get the number of rows affected")

	require.Equal(t, int64(1), rows, "rows returned is not correct")
}

func InsertSlotifyGroup(t *testing.T, db *sql.DB) api.SlotifyGroup {
	name := gofakeit.ProductName()

	var err error
	var res sql.Result
	// If deadlock sql error, retry upto 3 times, we had a few tests failing because of this.
	err = retry.Do(func() error {
		res, err = db.Exec("INSERT INTO SlotifyGroup (name) VALUES (?)", name)
		if err != nil {
			if database.IsDeadlockSQLError(err) {
				return fmt.Errorf("deadlock attempting to insert slotify group: %w", err)
			}

			return retry.Unrecoverable(err)
		}

		return nil
	}, retry.Attempts(3), retry.Delay(2*time.Second))

	require.NoError(t, err, "db insert slotifyGroup failed")

	rows, err := res.RowsAffected()
	require.Equal(t, int64(1), rows, "rows affected after insert is 1")
	require.NoError(t, err, "failed to get rows affected")

	id, err := res.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return api.SlotifyGroup{
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

	db := NewDB(t, ctx)

	server, err := api.NewServerWithContext(ctx, db,
		api.WithNotInitMSALClient(), api.WithNotificationService(notificationService))

	require.NoError(t, err, "error creating server ")
	require.NotNil(t, db, "server cannot be nil")

	return db, server
}

func NewDB(t *testing.T,
	ctx context.Context,
) *database.Database {
	db, err := database.NewDatabaseWithContext(ctx)
	require.NoError(t, err, "error creating database handle")
	require.NotNil(t, db, "db handle cannot be nil")

	return db
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
