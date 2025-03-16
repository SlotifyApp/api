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
			ctx = context.WithValue(ctx, api.RequestIDCtxKey{}, uuid.NewString())
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
	u := testutil.InsertUser(t, db)
	testutil.AddUserToSlotifyGroup(t, db, u.Id, slotifyGroupInserted.Id)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		slotifyGroupID   uint32
		testMsg          string
	}{
		"deleting slotifyGroup that doesn't exist": {
			expectedRespBody: "You are not a member of the group, you cannot delete it.",
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
			ctx := context.WithValue(req.Context(), api.UserIDCtxKey{}, u.Id)

			req = req.WithContext(ctx)

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
			ctx := context.WithValue(req.Context(), api.RequestIDCtxKey{}, uuid.NewString())
			req = req.WithContext(ctx)

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

	var zerothPage uint32

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
		t.Log("TestSlotifyGroup_GetSlotifyGroupsSlotifyGroupIDUsers range tests")
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf(
					"/api/slotify-groups/%d/users?pageToken=%d&limit=%d",
					tt.slotifyGroupID,
					zerothPage,
					testutil.PageLimit,
				),
				nil,
			)

			req.Header.Set(api.ReqHeader, uuid.NewString())

			server.GetAPISlotifyGroupsSlotifyGroupIDUsers(rr,
				req,
				tt.slotifyGroupID,
				api.GetAPISlotifyGroupsSlotifyGroupIDUsersParams{
					PageToken: &zerothPage,
					Limit:     testutil.PageLimit,
				},
			)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respBody struct {
					Users         api.Users `json:"users"`
					NextPageToken int       `json:"nextPageToken"`
				}
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into Users struct")
				require.Equal(t, tt.expectedRespBody, respBody.Users, tt.testMsg)
			} else {
				var respBody string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			}
		})
	}
	t.Run("pagination", func(t *testing.T) {
		t.Log("TestSlotifyGroup_GetSlotifyGroupsSlotifyGroupIDUsers pagination")
		groupForPagination := testutil.InsertSlotifyGroup(t, db)
		for range 11 {
			newUser := testutil.InsertUser(t, db)
			testutil.AddUserToSlotifyGroup(t, db, newUser.Id, groupForPagination.Id)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf(
				"/api/slotify-groups/%d/users?pageToken=%d&limit=%d",
				groupForPagination.Id,
				zerothPage,
				testutil.PageLimit,
			),
			nil,
		)
		req.Header.Set(api.ReqHeader, uuid.NewString())
		server.GetAPISlotifyGroupsSlotifyGroupIDUsers(
			rr,
			req,
			groupForPagination.Id,
			api.GetAPISlotifyGroupsSlotifyGroupIDUsersParams{
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
		var firstPageResp struct {
			Users         api.Users `json:"users"`
			NextPageToken int       `json:"nextPageToken"`
		}
		err = json.NewDecoder(rr.Result().Body).Decode(&firstPageResp)
		require.NoError(t, err)
		require.Len(t, firstPageResp.Users, 10, "first page should return 10 users")
		require.NotEqual(t, -1, firstPageResp.NextPageToken, "nextPageToken should be set if more users exist")

		rr2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf(
				"/api/slotify-groups/%d/users?pageToken=%d&limit=%d",
				groupForPagination.Id,
				firstPageResp.NextPageToken,
				testutil.PageLimit,
			),
			nil,
		)
		req2.Header.Set(api.ReqHeader, uuid.NewString())

		nextToken := uint32(firstPageResp.NextPageToken)
		server.GetAPISlotifyGroupsSlotifyGroupIDUsers(
			rr2,
			req2,
			groupForPagination.Id,
			api.GetAPISlotifyGroupsSlotifyGroupIDUsersParams{
				PageToken: &nextToken,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr2.Result().StatusCode)
		var secondPageResp struct {
			Users         api.Users `json:"users"`
			NextPageToken int       `json:"nextPageToken"`
		}
		err = json.NewDecoder(rr2.Result().Body).Decode(&secondPageResp)
		require.NoError(t, err)
		require.Len(t, secondPageResp.Users, 1, "second page should return 1 user")

		// no overlapping users between pages.
		firstPageIDs := make(map[uint32]bool)
		for _, u := range firstPageResp.Users {
			firstPageIDs[u.Id] = true
		}
		for _, u := range secondPageResp.Users {
			require.False(t, firstPageIDs[u.Id], "user id %d appears in both pages", u.Id)
		}
	})
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

	var zerothPage uint32

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
		t.Log("TestSlotifyGroup_GetAPISlotifyGroupsMe range tests")
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/slotify-groups/me?pageToken=%d&limit=%d", zerothPage, testutil.PageLimit),
				nil,
			)

			ctx := context.WithValue(req.Context(), api.UserIDCtxKey{}, tt.userID)
			ctx = context.WithValue(ctx, api.RequestIDCtxKey{}, uuid.NewString())

			req = req.WithContext(ctx)

			req.Header.Set(api.ReqHeader, uuid.NewString())

			server.GetAPISlotifyGroupsMe(rr, req, api.GetAPISlotifyGroupsMeParams{
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			})

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respBody struct {
					SlotifyGroup  api.SlotifyGroups `json:"slotifyGroups"`
					NextPageToken int               `json:"nextPageToken"`
				}
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into Users struct")
				require.Equal(t, tt.expectedRespBody, respBody.SlotifyGroup, tt.testMsg)
			} else {
				var respBody string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
				require.NoError(t, err, "response cannot be decoded into string")
				require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
			}
		})
	}

	t.Run("pagination", func(t *testing.T) {
		t.Log("TestSlotifyGroup_GetAPISlotifyGroupsMe pagination")
		// new user to prevent interference with previous tests
		user3 := testutil.InsertUser(t, db)
		for range 11 {
			g := testutil.InsertSlotifyGroup(t, db)
			testutil.AddUserToSlotifyGroup(t, db, user3.Id, g.Id)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/slotify-groups/me?pageToken=%d&limit=%d", zerothPage, testutil.PageLimit),
			nil,
		)
		ctx := context.WithValue(req.Context(), api.UserIDCtxKey{}, user3.Id)
		ctx = context.WithValue(ctx, api.RequestIDCtxKey{}, uuid.NewString())
		req = req.WithContext(ctx)
		req.Header.Set(api.ReqHeader, uuid.NewString())
		server.GetAPISlotifyGroupsMe(rr,
			req,
			api.GetAPISlotifyGroupsMeParams{
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
		var firstPageResp struct {
			SlotifyGroups api.SlotifyGroups `json:"slotifyGroups"`
			NextPageToken int               `json:"nextPageToken"`
		}
		err = json.NewDecoder(rr.Result().Body).Decode(&firstPageResp)
		require.NoError(t, err)
		require.Len(t, firstPageResp.SlotifyGroups, 10, "first page should return 10 groups")
		require.NotEqual(t, -1, firstPageResp.NextPageToken, "nextPageToken should be set if more groups exist")

		rr2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/slotify-groups/me?pageToken=%d&limit=%d", firstPageResp.NextPageToken, testutil.PageLimit),
			nil,
		)
		ctx2 := context.WithValue(req2.Context(), api.UserIDCtxKey{}, user3.Id)
		ctx2 = context.WithValue(ctx2, api.RequestIDCtxKey{}, uuid.NewString())
		req2 = req2.WithContext(ctx2)
		req2.Header.Set(api.ReqHeader, uuid.NewString())

		nextToken := uint32(firstPageResp.NextPageToken)
		server.GetAPISlotifyGroupsMe(rr2,
			req2,
			api.GetAPISlotifyGroupsMeParams{
				PageToken: &nextToken,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr2.Result().StatusCode)
		var secondPageResp struct {
			SlotifyGroups api.SlotifyGroups `json:"slotifyGroups"`
			NextPageToken int               `json:"nextPageToken"`
		}
		err = json.NewDecoder(rr2.Result().Body).Decode(&secondPageResp)
		require.NoError(t, err)
		require.Len(t, secondPageResp.SlotifyGroups, 1, "second page should return 1 group")
	})
}
