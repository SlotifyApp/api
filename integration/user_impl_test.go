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
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/SlotifyApp/slotify-backend/testutil"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestUser_GetUsersUserID(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	// Setup
	insertedUser := testutil.InsertUser(t, db)

	tests := map[string]struct {
		httpStatus   int
		userID       uint32
		expectedBody any
		testMsg      string
	}{
		"user does not exist": {
			httpStatus:   http.StatusNotFound,
			userID:       100000,
			expectedBody: "user api: user with id(100000) doesn't exist",
			testMsg:      "empty array is returned when slotify group does not exist",
		},
		"user exists": {
			httpStatus:   http.StatusOK,
			userID:       insertedUser.Id,
			expectedBody: insertedUser,
			testMsg:      "correctly got existing user",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d", tt.userID), nil)

			req.Header.Set(api.ReqHeader, uuid.NewString())

			server.GetAPIUsersUserID(rr, req, tt.userID)

			if tt.httpStatus == http.StatusOK {
				var user api.User
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&user)
				require.NoError(t, err, "response body can be decoded into string")
				require.Equal(t, tt.expectedBody, user, "got correct body")
			} else {
				var errMsg string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
				require.NoError(t, err, "response body can be decoded into string")
				require.Equal(t, tt.expectedBody, errMsg, "json body has correct error message")
			}

			testutil.OpenAPIValidateTest(t, rr, req)
		})
	}
}

func TestUser_PostUsers(t *testing.T) {
	t.Parallel()

	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	userCreate := api.UserCreate{
		Email:     openapi_types.Email(gofakeit.Email()),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
	}
	insertedUser := testutil.InsertUser(t, db)

	tests := map[string]struct {
		httpStatus   int
		userCreate   api.UserCreate
		expectedBody any
		testMsg      string
	}{
		"new user insert": {
			httpStatus: http.StatusCreated,
			userCreate: userCreate,
			testMsg:    "new user is successfully inserted",
		},
		"attempt to insert user with email that already exists": {
			httpStatus:   http.StatusBadRequest,
			expectedBody: fmt.Sprintf("user with email %s already exists", insertedUser.Email),
			userCreate: api.UserCreate{
				// Use same email but other fields don't matter
				Email:     insertedUser.Email,
				FirstName: gofakeit.FirstName(),
				LastName:  gofakeit.LastName(),
			},
			testMsg: "correctly detect email has already been used",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.userCreate)
			require.NoError(t, err, "could not marshal json req body user")

			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")

			req.Header.Set(api.ReqHeader, uuid.NewString())

			server.PostAPIUsers(rr, req)
			// Reset the request body for openapi validate
			req.Body = io.NopCloser(bytes.NewBuffer(body))

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusCreated {
				var user api.User
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&user)
				require.NoError(t, err, "response body can be decoded into string")

				require.Equal(t, tt.userCreate.Email, user.Email, "email is correct")
				require.Equal(t, tt.userCreate.FirstName, user.FirstName, "firstname is correct")
				require.Equal(t, tt.userCreate.LastName, user.LastName, "last name is correct")
			} else {
				var errMsg string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
				require.NoError(t, err, "response body can be decoded into string")
				require.Equal(t, tt.expectedBody, errMsg, tt.testMsg)
			}

			testutil.OpenAPIValidateTest(t, rr, req)
		})
	}
}

func TestUser_GetUsers(t *testing.T) {
	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	// Setup
	fakeLastName := gofakeit.LastName() + "blah"
	insertedUser := testutil.InsertUser(t, db)
	insertedUserEmail := string(insertedUser.Email)
	insertedUser2 := testutil.InsertUser(t, db, testutil.WithFirstName(insertedUser.FirstName))
	insertedUser3 := testutil.InsertUser(t, db, testutil.WithLastName(insertedUser.LastName))

	tests := map[string]struct {
		httpStatus   int
		expectedBody any
		testMsg      string
		route        string
		params       api.GetAPIUsersParams
	}{
		"get existing user by email that exists": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{insertedUser},
			testMsg:      "successfully got user by email",
			route: fmt.Sprintf(
				"email=%s",
				url.QueryEscape(string(insertedUser.Email)),
			),
			params: api.GetAPIUsersParams{
				Email: &insertedUserEmail,
			},
		},
		"get existing user by first name": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{insertedUser, insertedUser2},
			testMsg:      "successfully got users by first name",
			route: fmt.Sprintf(
				"name=%s",
				url.QueryEscape(insertedUser.FirstName),
			),
			params: api.GetAPIUsersParams{
				Name: &insertedUser.FirstName,
			},
		},
		"get existing user by last name": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{insertedUser, insertedUser3},
			testMsg:      "successfully got users by last name",
			route: fmt.Sprintf(
				"name=%s",
				url.QueryEscape(insertedUser.LastName),
			),
			params: api.GetAPIUsersParams{
				Name: &insertedUser.LastName,
			},
		},
		"get users by non-existent query params": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{},
			testMsg:      "successfully got empty array of users when users don't exist by query params",
			route: fmt.Sprintf(
				"name=%s",
				url.QueryEscape(fakeLastName),
			),
			params: api.GetAPIUsersParams{
				Name: &fakeLastName,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/users?%s", tt.route),
				nil,
			)
			ctx := context.WithValue(req.Context(), api.RequestIDCtxKey{}, uuid.NewString())
			req = req.WithContext(ctx)

			req.Header.Add("Content-Type", "application/json")

			server.GetAPIUsers(rr, req, tt.params)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respUsers api.Users
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
				require.NoError(t, err, "response body can be decoded into a User")

				require.Equal(t, tt.expectedBody, respUsers, tt.testMsg)
			} else {
				var errMsg string
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
				require.NoError(t, err, "response body can be decoded into a string")
				require.Equal(t, tt.expectedBody, errMsg, tt.testMsg)
			}
		})
	}

	// Don't want to assert every user in a var, so separate test
	t.Run("route with no query params returns error", func(t *testing.T) {
		require.NoError(t, err, "could not begin transaction")

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/users",
			nil,
		)
		req.Header.Add("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), api.RequestIDCtxKey{}, uuid.NewString())
		req = req.WithContext(ctx)
		require.NoError(t, err, "failed to commit transaction")

		server.GetAPIUsers(rr, req, api.GetAPIUsersParams{})
		testutil.OpenAPIValidateTest(t, rr, req)

		var body string
		require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
		err = json.NewDecoder(rr.Result().Body).Decode(&body)
		require.NoError(t, err, "response body can be decoded into a string")

		require.Equal(t, "please provide at least email or name",
			body, "get api users with no query params returns error")
	})
}

func TestUser_DeleteUsersUserID(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, t.Context())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	fakeUserID := uint32(10000)
	tests := map[string]struct {
		httpStatus   int
		expectedBody any
		testMsg      string
		userID       uint32
	}{
		"delete user that doesn't exist": {
			httpStatus:   http.StatusBadRequest,
			userID:       fakeUserID,
			expectedBody: fmt.Sprintf("user api: user with id(%d) doesn't exist", fakeUserID),
			testMsg:      "correct error response when deleting user that does not exist",
		},

		"delete user that exists": {
			httpStatus:   http.StatusBadRequest,
			userID:       fakeUserID,
			expectedBody: fmt.Sprintf("user api: user with id(%d) doesn't exist", fakeUserID),
			testMsg:      "correct error response when deleting user that does not exist",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%d", tt.userID), nil)

			req.Header.Set(api.ReqHeader, uuid.NewString())
			req.Header.Add("Content-Type", "application/json")

			oldCount := testutil.GetCount(t, db, "User")

			server.DeleteAPIUsersUserID(rr, req, tt.userID)
			testutil.OpenAPIValidateTest(t, rr, req)

			require.Equal(t, tt.httpStatus, rr.Result().StatusCode)

			if tt.httpStatus == http.StatusOK {
				newCount := testutil.GetCount(t, db, "User")
				require.Equal(t, newCount, oldCount-1, "user deleted from db")
			}

			var actualBody string
			err = json.NewDecoder(rr.Result().Body).Decode(&actualBody)
			require.NoError(t, err, "decode returns error")

			require.Equal(t, tt.expectedBody, actualBody, tt.testMsg)
		})
	}
}
