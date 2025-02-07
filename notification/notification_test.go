package notification_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/notification"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/stretchr/testify/require"
)

func Test_SSERegisterUserClient(t *testing.T) {
	t.Parallel()

	l := testutil.NewLogger(t)

	var userID1 uint32 = 5

	// Test 2 set up
	w1 := httptest.NewRecorder()
	registeredClientsMap1 := make(map[uint32]notification.ClientSet)
	registeredClientsMap1[userID1] = make(notification.ClientSet)
	registeredClientsMap1[userID1][w1] = struct{}{}

	// Test 3 set up
	var userID2 uint32 = 10
	w2 := httptest.NewRecorder()
	w3 := httptest.NewRecorder()
	registeredClientsMap2 := make(map[uint32]notification.ClientSet)
	registeredClientsMap2[userID2] = make(notification.ClientSet)
	registeredClientsMap2[userID2][w2] = struct{}{}
	registeredClientsMap2[userID2][w3] = struct{}{}

	tests := map[string]struct {
		clientToAdd     []http.ResponseWriter
		userID          uint32
		testMsg         string
		expectedError   error
		expectedClients map[uint32]notification.ClientSet
	}{
		"register nil client": {
			clientToAdd:     nil,
			testMsg:         "registering nil client returns error",
			userID:          userID1,
			expectedError:   notification.ErrNotifClientNil,
			expectedClients: nil,
		},
		"register client": {
			clientToAdd:     []http.ResponseWriter{w1},
			testMsg:         "registering 1 client is correct",
			userID:          userID1,
			expectedError:   nil,
			expectedClients: registeredClientsMap1,
		},
		"register client that already exists does nothing": {
			clientToAdd:     []http.ResponseWriter{w1, w1},
			testMsg:         "registering 1 client is correct",
			userID:          userID1,
			expectedError:   nil,
			expectedClients: registeredClientsMap1,
		},
		"register multiple clients": {
			clientToAdd:     []http.ResponseWriter{w2, w3},
			testMsg:         "registering nil client returns error",
			userID:          userID2,
			expectedError:   nil,
			expectedClients: registeredClientsMap2,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			sseNotifService := notification.NewSSENotificationService()
			// Register clients
			for _, w := range tt.clientToAdd {
				receivedErr := sseNotifService.RegisterUserClient(l, tt.userID, w)

				if tt.expectedError != nil {
					require.ErrorIs(t, receivedErr, tt.expectedError, tt.testMsg)
				} else {
					require.NoError(t, receivedErr,
						"registering valid clients doesn't return an error")
				}
			}

			// tests specific to registering valid clients
			if tt.expectedError == nil {
				clients := sseNotifService.GetUserClients()
				require.True(t, reflect.DeepEqual(tt.expectedClients, clients),
					"registered clients are equal")
			}
		})
	}

	// Test that multiple clients for multiple users are correct

	// Test 4 set up
	allRegisteredClients := make(map[uint32]notification.ClientSet)
	allRegisteredClients[userID1] = make(notification.ClientSet)
	allRegisteredClients[userID1][w1] = struct{}{}
	allRegisteredClients[userID2] = make(notification.ClientSet)
	allRegisteredClients[userID2][w2] = struct{}{}
	allRegisteredClients[userID2][w3] = struct{}{}

	sseNotifService := notification.NewSSENotificationService()

	require.NoError(t, sseNotifService.RegisterUserClient(l, userID1, w1), "registering user client is not error")
	require.NoError(t, sseNotifService.RegisterUserClient(l, userID2, w2), "registering user client is not error")
	require.NoError(t, sseNotifService.RegisterUserClient(l, userID2, w3), "registering user client is not error")

	clients := sseNotifService.GetUserClients()
	require.True(t, reflect.DeepEqual(clients, allRegisteredClients),
		"multiple clients for multiple users registered correctly")
}

func Test_SSEDeleteUserConn(t *testing.T) {
	t.Parallel()

	l := testutil.NewLogger(t)

	// Test 1 set up
	var userID1 uint32 = 1
	var userID2 uint32 = 2
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	w3 := httptest.NewRecorder()

	// test1: remove userID1's w1 client
	expectedClients1 := make(map[uint32]notification.ClientSet)
	expectedClients1[userID1] = make(notification.ClientSet)
	expectedClients1[userID2] = make(notification.ClientSet)

	expectedClients1[userID2][w2] = struct{}{}
	expectedClients1[userID2][w3] = struct{}{}

	// test2: remove existing user's non-existent client
	expectedClients2 := make(map[uint32]notification.ClientSet)
	expectedClients2[userID1] = make(notification.ClientSet)
	expectedClients2[userID2] = make(notification.ClientSet)
	expectedClients2[userID1] = make(notification.ClientSet)

	expectedClients2[userID1][w1] = struct{}{}
	expectedClients2[userID2][w2] = struct{}{}
	expectedClients2[userID2][w3] = struct{}{}

	// test3: remove user2's w2 conn
	expectedClients3 := make(map[uint32]notification.ClientSet)
	expectedClients3[userID1] = make(notification.ClientSet)
	expectedClients3[userID2] = make(notification.ClientSet)
	expectedClients3[userID1] = make(notification.ClientSet)

	expectedClients3[userID1][w1] = struct{}{}
	expectedClients3[userID2][w3] = struct{}{}

	// test4: remove user2's w3 conn
	expectedClients4 := make(map[uint32]notification.ClientSet)
	expectedClients4[userID1] = make(notification.ClientSet)
	expectedClients4[userID2] = make(notification.ClientSet)
	expectedClients4[userID1] = make(notification.ClientSet)

	expectedClients4[userID1][w1] = struct{}{}
	expectedClients4[userID2][w2] = struct{}{}

	// test5: remove non-existing user's existent client
	expectedClients5 := make(map[uint32]notification.ClientSet)
	expectedClients5[userID1] = make(notification.ClientSet)
	expectedClients5[userID2] = make(notification.ClientSet)
	expectedClients5[userID1] = make(notification.ClientSet)

	expectedClients5[userID1][w1] = struct{}{}
	expectedClients5[userID2][w2] = struct{}{}
	expectedClients5[userID2][w3] = struct{}{}

	tests := map[string]struct {
		clientToRemove  http.ResponseWriter
		userID          uint32
		testMsg         string
		expectedClients map[uint32]notification.ClientSet
	}{
		"remove user's only client": {
			clientToRemove:  w1,
			testMsg:         "removing only user's client is successful",
			userID:          userID1,
			expectedClients: expectedClients1,
		},
		"remove user's non-existent client": {
			clientToRemove:  httptest.NewRecorder(),
			testMsg:         "removing user's non-existent client doesn't do anything",
			userID:          userID1,
			expectedClients: expectedClients2,
		},
		"remove one of a user's client": {
			clientToRemove:  w2,
			testMsg:         "removing user's legit client is successful",
			userID:          userID2,
			expectedClients: expectedClients3,
		},
		"remove another one of a user's client": {
			clientToRemove:  w3,
			testMsg:         "removing user's legit client is successful",
			userID:          userID2,
			expectedClients: expectedClients4,
		},

		"remove non-existing user's client does nothing": {
			clientToRemove:  w3,
			testMsg:         "removing user's legit client is successful",
			userID:          1000,
			expectedClients: expectedClients5,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			sseNotifService := notification.NewSSENotificationService()
			sseNotifService.RegisterUserClient(l, userID1, w1)
			sseNotifService.RegisterUserClient(l, userID2, w2)
			sseNotifService.RegisterUserClient(l, userID2, w3)

			sseNotifService.DeleteUserConn(l, tt.userID, tt.clientToRemove)

			clients := sseNotifService.GetUserClients()

			require.True(t, reflect.DeepEqual(tt.expectedClients, clients),
				tt.testMsg)
		})
	}
}

func Test_SSESendNotification(t *testing.T) {
	t.Parallel()
	l := testutil.NewLogger(t)

	var notifID int64 = 1
	var userID uint32 = 1
	w1 := httptest.NewRecorder()

	mockNotificationDB := testutil.NewMockNotificationDatabase(1, notifID)

	sseNotificationService := notification.NewSSENotificationService()
	require.NoError(t, sseNotificationService.RegisterUserClient(l, userID, w1),
		"registering client should not return erro")

	created := time.Now()
	message := "This is a notification"
	notifParams := database.CreateNotificationParams{
		Message: message,
		Created: created,
	}

	err := sseNotificationService.SendNotification(context.Background(), l, mockNotificationDB, userID, notifParams)
	require.NoError(t, err,
		"send notification should execute successfully and produce no error")

	present := sseNotificationService.GetUserClients()[userID][w1]
	require.Equal(t, struct{}{}, present, "client was correctly registered for user")

	resp := w1.Result()
	require.True(t, w1.Flushed, "notification was flushed to body")
	var body []byte
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	expectedData := fmt.Sprintf(
		"event: calendar_notification\ndata: {\"id\":%d,\"message\":\"%s\",\"created\":\"%s\"}\n\n",
		notifID, message, created.Format(time.RFC3339Nano))

	require.Equal(t, expectedData, string(body), "body of event stream is correct")
}
