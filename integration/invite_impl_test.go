package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/mocks"
	"github.com/SlotifyApp/slotify-backend/testutil"
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
			req = req.WithContext(ctx)

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
