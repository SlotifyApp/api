package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/mocks"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
			testMsg: "created group successfully",
		},
		"slotifyGroup already exists": {
			httpStatus: http.StatusBadRequest,
			slotifyGroupBody: api.SlotifyGroupCreate{
				Name: insertedSlotifyGroup.Name,
			},
			expectedRespBody: fmt.Sprintf("slotifyGroup with name %s already exists", insertedSlotifyGroup.Name),
			testMsg:          "slotify group that already exists has correct message",
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

			req.Header.Set(api.ReqHeader, uuid.NewString())

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

			req.Header.Set(api.ReqHeader, uuid.NewString())

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

			req.Header.Set(api.ReqHeader, uuid.NewString())

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
			httpStatus:       http.StatusNotFound,
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

			req.Header.Set(api.ReqHeader, uuid.NewString())

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

			req.Header.Set(api.ReqHeader, uuid.NewString())

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
