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
	"time"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/mocks"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestInvites_PostInvites(t *testing.T) {
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

	fromUser := testutil.InsertUser(t, db)
	toUser := testutil.InsertUser(t, db)

	// Setup
	slotifyGroup := testutil.InsertSlotifyGroup(t, db)

	testutil.AddUserToSlotifyGroup(t, db, fromUser.Id, slotifyGroup.Id)

	// test 1 setup
	inviteBody1 := api.PostAPIInvitesJSONRequestBody{
		CreatedAt:      time.Now(),
		ExpiryDate:     openapi_types.Date{Time: time.Now().Add(time.Hour * 24)},
		Message:        "Hey, this is the invite message",
		SlotifyGroupID: slotifyGroup.Id,
		ToUserID:       toUser.Id,
	}

	expectedInviteGroup := api.InvitesGroup{
		CreatedAt:         inviteBody1.CreatedAt,
		ExpiryDate:        inviteBody1.ExpiryDate,
		FromUserEmail:     fromUser.Email,
		FromUserFirstName: fromUser.FirstName,
		FromUserLastName:  fromUser.LastName,
		Message:           inviteBody1.Message,
		Status:            api.InviteStatusPending,
		ToUserEmail:       toUser.Email,
		ToUserFirstName:   toUser.FirstName,
		ToUserLastName:    toUser.LastName,
	}

	tests := map[string]struct {
		httpStatus          int
		inviteBody          api.PostAPIInvitesJSONRequestBody
		expectedRespBody    api.InvitesGroup
		expectedErrRespBody string
		testMsg             string
	}{
		"invite created successfully": {
			httpStatus:       http.StatusCreated,
			inviteBody:       inviteBody1,
			expectedRespBody: expectedInviteGroup,
			testMsg:          "created invite correctly",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.inviteBody)
			require.NoError(t, err, "could not marshal json req body inviteBody")
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/invites", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), api.UserIDCtxKey{}, fromUser.Id)
			ctx = context.WithValue(ctx, api.RequestIDCtxKey{}, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set(api.ReqHeader, uuid.NewString())

			server.PostAPIInvites(rr, req)

			if tt.httpStatus == http.StatusCreated {
				var body api.InvitesGroup
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				require.NoError(t, json.NewDecoder(rr.Result().Body).Decode(&body),
					"failed to decode body")

				require.Equal(t, tt.expectedRespBody.FromUserEmail, body.FromUserEmail, "FromUserEmail was equal")

				require.Equal(t, tt.expectedRespBody.FromUserLastName, body.FromUserLastName, "FromUserLastName was equal")
				require.Equal(t, tt.expectedRespBody.Message, body.Message, "Message was equal")
				require.Equal(t, tt.expectedRespBody.Status, body.Status, "Status was equal")
				require.Equal(t, tt.expectedRespBody.ToUserEmail, body.ToUserEmail, "ToUserEmail was equal")
				require.Equal(t, tt.expectedRespBody.ToUserFirstName, body.ToUserFirstName, "ToUserFirstName was equal")
				require.Equal(t, tt.expectedRespBody.ToUserLastName, body.ToUserLastName, "ToUserLastName was equal")
				require.WithinDuration(t, tt.expectedRespBody.CreatedAt, body.CreatedAt, time.Second,
					"ExpiryDate mismatch")

				require.Equal(t, tt.expectedRespBody.ExpiryDate.Format(time.DateOnly),
					body.ExpiryDate.Format(time.DateOnly), "expiryDate is equal")
			} else {
				var errMsg string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				require.NoError(t, json.NewDecoder(rr.Result().Body).Decode(&errMsg),
					"response cannot be decoded into string")
				require.Equal(t, tt.expectedErrRespBody, errMsg, tt.testMsg)
			}

			req.Body = io.NopCloser(bytes.NewBuffer(body))
			testutil.OpenAPIValidateTest(t, rr, req)
		})
	}
}

func TestAPISlotifyGroupsSlotifyGroupIDInvites(t *testing.T) {
	var err error
	datab, server := testutil.NewServerAndDB(t, t.Context())
	db := datab.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	testUser := testutil.InsertUser(t, db)
	testUser2 := testutil.InsertUser(t, db)
	testGroup := testutil.InsertSlotifyGroup(t, db)
	testutil.AddUserToSlotifyGroup(t, db, testUser.Id, testGroup.Id)
	testutil.AddUserToSlotifyGroup(t, db, testUser2.Id, testGroup.Id)
	var zerothPage uint32

	t.Run("no invites", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/slotify-groups/%d/invites?pageToken=%d&limit=%d", testGroup.Id, zerothPage, testutil.PageLimit),
			nil,
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set(api.ReqHeader, uuid.NewString())
		req = req.WithContext(context.WithValue(req.Context(), api.UserIDCtxKey{}, testUser.Id))
		server.GetAPISlotifyGroupsSlotifyGroupIDInvites(
			rr,
			req,
			testGroup.Id,
			api.GetAPISlotifyGroupsSlotifyGroupIDInvitesParams{
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)

		var resp struct {
			Invites       []database.ListInvitesByGroupRow `json:"invites"`
			NextPageToken int                              `json:"nextPageToken,omitempty"`
		}
		err = json.NewDecoder(rr.Result().Body).Decode(&resp)
		require.NoError(t, err)
		require.Empty(t, resp.Invites, "should return no invites")
		require.Equal(t, 0, resp.NextPageToken, "nextPageToken should be 0 when no invites")
	})

	t.Run("less than limit invites", func(t *testing.T) {
		for range 5 {
			testutil.InsertInvite(t, db, testUser2, testUser, testGroup.Id)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/slotify-groups/%d/invites?pageToken=%d&limit=%d", testGroup.Id, zerothPage, testutil.PageLimit),
			nil,
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set(api.ReqHeader, uuid.NewString())
		req = req.WithContext(context.WithValue(req.Context(), api.UserIDCtxKey{}, testUser.Id))
		server.GetAPISlotifyGroupsSlotifyGroupIDInvites(
			rr,
			req,
			testGroup.Id,
			api.GetAPISlotifyGroupsSlotifyGroupIDInvitesParams{
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)

		var resp struct {
			Invites       []database.ListInvitesByGroupRow `json:"invites"`
			NextPageToken int                              `json:"nextPageToken,omitempty"`
		}
		err = json.NewDecoder(rr.Result().Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Invites, 5, "should return 5 invites")
		require.Equal(t, 0, resp.NextPageToken, "nextPageToken should be 0 when invites are less than limit")
	})

	t.Run("pagination", func(t *testing.T) {
		for range 6 {
			testutil.InsertInvite(t, db, testUser2, testUser, testGroup.Id)
		}
		// first page request.
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/slotify-groups/%d/invites?pageToken=%d&limit=%d", testGroup.Id, zerothPage, testutil.PageLimit),
			nil,
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set(api.ReqHeader, uuid.NewString())
		req = req.WithContext(context.WithValue(req.Context(), api.UserIDCtxKey{}, testUser.Id))
		server.GetAPISlotifyGroupsSlotifyGroupIDInvites(
			rr,
			req,
			testGroup.Id,
			api.GetAPISlotifyGroupsSlotifyGroupIDInvitesParams{
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)

		var resp struct {
			Invites       []database.ListInvitesByGroupRow `json:"invites"`
			NextPageToken int                              `json:"nextPageToken,omitempty"`
		}
		err = json.NewDecoder(rr.Result().Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Invites, 10, "first page should return 10 invites")
		require.NotEqual(t, -1, resp.NextPageToken, "nextPageToken should be non -1 if more invites exist")

		// Second page request using the returned nextPageToken.
		rr2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf(
				"/api/slotify-groups/%d/invites?pageToken=%d&limit=%d",
				testGroup.Id,
				resp.NextPageToken,
				testutil.PageLimit,
			),
			nil,
		)
		req2.Header.Add("Content-Type", "application/json")
		req2.Header.Set(api.ReqHeader, uuid.NewString())
		req2 = req2.WithContext(context.WithValue(req2.Context(), api.UserIDCtxKey{}, testUser.Id))

		nextToken := uint32(resp.NextPageToken)
		server.GetAPISlotifyGroupsSlotifyGroupIDInvites(
			rr2,
			req2,
			testGroup.Id,
			api.GetAPISlotifyGroupsSlotifyGroupIDInvitesParams{
				PageToken: &nextToken,
				Limit:     testutil.PageLimit,
			},
		)
		require.Equal(t, http.StatusOK, rr2.Result().StatusCode)

		var resp2 struct {
			Invites       []database.ListInvitesByGroupRow `json:"invites"`
			NextPageToken int                              `json:"nextPageToken,omitempty"`
		}
		err = json.NewDecoder(rr2.Result().Body).Decode(&resp2)
		require.NoError(t, err)
		require.Len(t, resp2.Invites, 1, "second page should return the remaining invite")
		page1IDs := make(map[uint32]bool)
		for _, inv := range resp.Invites {
			page1IDs[inv.InviteID] = true
		}
		for _, inv := range resp2.Invites {
			require.False(t, page1IDs[inv.InviteID], "invite id %d appears in both pages", inv.InviteID)
		}
	})
}
