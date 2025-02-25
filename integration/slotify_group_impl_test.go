package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/mocks"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSlotifyGroup_GetSlotifyGroups(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	// For testing, we want the underlying db connection rather than the
	// sqlc queries.
	db := database.DB

	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	insertedSlotifyGroup := testutil.InsertSlotifyGroup(t, db)

	tests := map[string]struct {
		httpStatus            int
		slotifyGroupName      string
		expectedSlotifyGroups api.SlotifyGroups
		testMsg               string
		route                 string
	}{
		"slotifyGroup does not exist": {
			httpStatus:            http.StatusOK,
			slotifyGroupName:      "DoesntExist",
			expectedSlotifyGroups: api.SlotifyGroups{},
			testMsg:               "empty array is returned when slotifyGroup does not exist",
		},
		"slotifyGroup does exist": {
			httpStatus:            http.StatusOK,
			slotifyGroupName:      insertedSlotifyGroup.Name,
			expectedSlotifyGroups: api.SlotifyGroups{insertedSlotifyGroup},
			testMsg:               "correct array is returned when slotifyGroup exists",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			params := api.GetAPISlotifyGroupsParams{
				Name: &tt.slotifyGroupName,
			}
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/api/slotify-groups?name=%s", url.QueryEscape(*params.Name)), nil)

			server.GetAPISlotifyGroups(rr, req, params)

			var slotifyGroups api.SlotifyGroups
			require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&slotifyGroups)
			require.NoError(t, err, "response cannot be decoded into SlotifyGroups struct")
			require.Equal(t, tt.expectedSlotifyGroups, slotifyGroups, tt.testMsg)

			testutil.OpenAPIValidateTest(t, rr, req)
		})
	}
}

func TestSlotifyGroup_PostSlotifyGroups(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockNotifService := mocks.NewMockService(ctrl)

	mockNotifService.
		EXPECT().
		SendNotification(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	database, server := testutil.NewServerAndDB(t,
		t.Context(),
		testutil.WithNotificationService(mockNotifService))

	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})
	user := testutil.InsertUser(t, db)
	// Setup
	insertedSlotifyGroup := testutil.InsertSlotifyGroup(t, db)
	newSlotifyGroupName := gofakeit.ProductName()

	tests := map[string]struct {
		httpStatus       int
		slotifyGroupBody api.SlotifyGroupCreate
		slotifyGroupName string
		expectedRespBody any
		testMsg          string
	}{
		"slotifyGroup inserted correctly": {
			httpStatus:       http.StatusCreated,
			slotifyGroupName: newSlotifyGroupName,
			slotifyGroupBody: api.SlotifyGroupCreate{
				Name: newSlotifyGroupName,
			},
			testMsg: "slotifyGroup made successfully",
		},
		"slotifyGroup already exists": {
			httpStatus: http.StatusBadRequest,
			slotifyGroupBody: api.SlotifyGroupCreate{
				Name: insertedSlotifyGroup.Name,
			},
			expectedRespBody: fmt.Sprintf("slotifyGroup with name %s already exists", insertedSlotifyGroup.Name),
			testMsg:          "slotifyGroup that already exists",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.slotifyGroupBody)
			require.NoError(t, err, "could not marshal json req body slotifyGroup")
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/slotify-groups", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), api.UserIDCtxKey{}, user.Id)
			req = req.WithContext(ctx)

			server.PostAPISlotifyGroups(rr, req)

			if tt.httpStatus == http.StatusCreated {
				var slotifyGroup api.SlotifyGroup
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&slotifyGroup)
				require.NoError(t, err, "response cannot be decoded into SlotifyGroup struct")

				require.Equal(t, tt.slotifyGroupName, slotifyGroup.Name, tt.testMsg)
			} else {
				var errMsg string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, errMsg, tt.testMsg)
			}

			req.Body = io.NopCloser(bytes.NewBuffer(body))
			testutil.OpenAPIValidateTest(t, rr, req)
		})
	}
}

func TestSlotifyGroup_DeleteSlotifyGroupsSlotifyGroupID(t *testing.T) {
	t.Parallel()

	var err error

	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	slotifyGroupInserted := testutil.InsertSlotifyGroup(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		slotifyGroupID   uint32
		testMsg          string
	}{
		"deleting slotifyGroup that doesn't exist": {
			expectedRespBody: "slotifyGroup api: incorrect slotifyGroup id",
			httpStatus:       http.StatusBadRequest,
			slotifyGroupID:   10000,
			testMsg:          "slotifyGroup that doesn't exist, returns client error",
		}, "deleting slotifyGroup that exists": {
			expectedRespBody: "slotifyGroup api: slotifyGroup deleted successfully",
			httpStatus:       http.StatusOK,
			slotifyGroupID:   slotifyGroupInserted.Id,
			testMsg:          "deleting slotifyGroup that exists is successful",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/slotify-groups/%d", tt.slotifyGroupID), nil)

			server.DeleteAPISlotifyGroupsSlotifyGroupID(rr, req, tt.slotifyGroupID)

			testutil.OpenAPIValidateTest(t, rr, req)
			var errMsg string
			require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
			require.NoError(t, err, "response cannot be decoded into string")
			require.Equal(t, tt.expectedRespBody, errMsg, tt.testMsg)
		})
	}
}

func TestSlotifyGroup_GetSlotifyGroupsSlotifyGroupID(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	slotifyGroupInserted := testutil.InsertSlotifyGroup(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		slotifyGroupID   uint32
		testMsg          string
	}{
		"get slotifyGroup that exists": {
			expectedRespBody: slotifyGroupInserted,
			httpStatus:       http.StatusOK,
			slotifyGroupID:   slotifyGroupInserted.Id,
			testMsg:          "can get slotifyGroup that exists successfully",
		},
		"get slotifyGroup that doesn't exist": {
			expectedRespBody: "slotifyGroup api: slotifyGroup with id 1000 does not exist",
			httpStatus:       http.StatusNotFound,
			slotifyGroupID:   1000,
			testMsg:          "deleting slotifyGroup that doesn't exist is unsuccessful",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/slotify-groups/%d", tt.slotifyGroupID), nil)

			server.GetAPISlotifyGroupsSlotifyGroupID(rr, req, tt.slotifyGroupID)

			testutil.OpenAPIValidateTest(t, rr, req)
			if tt.httpStatus == http.StatusOK {
				var slotifyGroup api.SlotifyGroup
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&slotifyGroup)
				require.NoError(t, err, "response cannot be decoded into a slotifyGroup")
				require.Equal(t, tt.expectedRespBody, slotifyGroup, tt.testMsg)
			} else {
				var errMsg string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, errMsg, tt.testMsg)
			}
		})
	}
}

func TestSlotifyGroup_PostSlotifyGroupsSlotifyGroupIDUsersUserID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockNotifService := mocks.NewMockService(ctrl)

	mockNotifService.
		EXPECT().
		SendNotification(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	var err error
	database, server := testutil.NewServerAndDB(t,
		t.Context(),
		testutil.WithNotificationService(mockNotifService))

	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	slotifyGroupInserted := testutil.InsertSlotifyGroup(t, db)

	userInserted := testutil.InsertUser(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		slotifyGroupID   uint32
		userID           uint32
		testMsg          string
	}{
		"insert a user into a slotifyGroup": {
			expectedRespBody: api.SlotifyGroup{
				Id:   slotifyGroupInserted.Id,
				Name: slotifyGroupInserted.Name,
			},
			httpStatus:     http.StatusCreated,
			userID:         userInserted.Id,
			slotifyGroupID: slotifyGroupInserted.Id,
			testMsg:        "can add user to slotifyGroup where both exist successfully",
		},
		"insert a user into a non-existent slotifyGroup": {
			expectedRespBody: fmt.Sprintf("slotifyGroup api: slotifyGroup id(%d) or user id(%d) was invalid",
				1000, userInserted.Id),
			httpStatus:     http.StatusForbidden,
			userID:         userInserted.Id,
			slotifyGroupID: 1000,
			testMsg:        "cannot add user to slotifyGroup where slotifyGroup does not exist",
		},

		"insert an non-existent user into a slotifyGroup": {
			expectedRespBody: fmt.Sprintf("slotifyGroup api: slotifyGroup id(%d) or user id(%d) was invalid",
				slotifyGroupInserted.Id, 10000),
			httpStatus:     http.StatusForbidden,
			userID:         10000,
			slotifyGroupID: slotifyGroupInserted.Id,
			testMsg:        "cannot add user to slotifyGroup where user does not exist",
		},

		"user and slotifyGroup ids do not exist": {
			expectedRespBody: fmt.Sprintf("slotifyGroup api: slotifyGroup id(%d) or user id(%d) was invalid", 10000, 10000),
			httpStatus:       http.StatusForbidden,
			userID:           10000,
			slotifyGroupID:   10000,
			testMsg:          "cannot add user to slotifyGroup where neither exists",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/api/slotify-groups/%d/users/%d", tt.slotifyGroupID, tt.userID), nil)

			server.PostAPISlotifyGroupsSlotifyGroupIDUsersUserID(rr, req, tt.slotifyGroupID, tt.userID)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusCreated {
				var slotifyGroup api.SlotifyGroup
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&slotifyGroup)
				require.NoError(t, err, "response cannot be decoded into SlotifyGroup struct")
				require.Equal(t, tt.expectedRespBody, slotifyGroup, tt.testMsg)
			} else {
				var respBody string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			}
		})
	}
}

func TestSlotifyGroup_GetSlotifyGroupsSlotifyGroupIDUsers(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	slotifyGroupInserted := testutil.InsertSlotifyGroup(t, db)

	userInserted := testutil.InsertUser(t, db)
	userInserted2 := testutil.InsertUser(t, db)

	testutil.AddUserToSlotifyGroup(t, db, userInserted.Id, slotifyGroupInserted.Id)
	testutil.AddUserToSlotifyGroup(t, db, userInserted2.Id, slotifyGroupInserted.Id)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		slotifyGroupID   uint32
		testMsg          string
	}{
		"get members of a non-existing slotifyGroup": {
			expectedRespBody: "slotifyGroup api: slotifyGroup with id(10000) does not exist",
			httpStatus:       http.StatusForbidden,
			slotifyGroupID:   10000,
			testMsg:          "correct error returns when slotifyGroup doesn't exist",
		},
		"get members of an existing slotifyGroup": {
			expectedRespBody: api.Users{userInserted, userInserted2},
			httpStatus:       http.StatusOK,
			slotifyGroupID:   slotifyGroupInserted.Id,
			testMsg:          "correctly get members of a slotifyGroup",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/slotify-groups/%d/users", tt.slotifyGroupID), nil)

			server.GetAPISlotifyGroupsSlotifyGroupIDUsers(rr, req, tt.slotifyGroupID)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respBody api.Users
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into Users struct")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			} else {
				var respBody string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			}
		})
	}
}

func TestSlotifyGroup_GetAPISlotifyGroupsMe(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	user1 := testutil.InsertUser(t, db, testutil.WithEmail("blah@example.com"))

	insertedSlotifyGroup1 := testutil.InsertSlotifyGroup(t, db)
	insertedSlotifyGroup2 := testutil.InsertSlotifyGroup(t, db)
	user2 := testutil.InsertUser(t, db)
	testutil.AddUserToSlotifyGroup(t, db, user2.Id, insertedSlotifyGroup1.Id)
	testutil.AddUserToSlotifyGroup(t, db, user2.Id, insertedSlotifyGroup2.Id)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		testMsg          string
		userID           uint32
	}{
		"get slotifyGroups of user who has no slotifyGroups": {
			expectedRespBody: api.SlotifyGroups{},
			httpStatus:       http.StatusOK,
			testMsg:          "user who has no slotifyGroups returns empty list",
			userID:           user1.Id,
		},
		"get slotifyGroups of user who has many slotifyGroups": {
			expectedRespBody: api.SlotifyGroups{insertedSlotifyGroup1, insertedSlotifyGroup2},
			httpStatus:       http.StatusOK,
			testMsg:          "correctly get all of a user's slotifyGroups",
			userID:           user2.Id,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/slotify-groups/me", nil)

			ctx := context.WithValue(req.Context(), api.UserIDCtxKey{}, tt.userID)
			req = req.WithContext(ctx)

			server.GetAPISlotifyGroupsMe(rr, req)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respBody api.SlotifyGroups
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into Users struct")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			} else {
				var respBody string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			}
		})
	}
}
