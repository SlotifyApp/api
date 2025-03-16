package api_test

import (
	"bytes"
	"database/sql"
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
	insertedUser2 := testutil.InsertUser(t, db, testutil.WithFirstName(insertedUser.FirstName))
	insertedUser3 := testutil.InsertUser(t, db, testutil.WithLastName(insertedUser.LastName))

	var zerothPage uint32

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
				"?email=%s&pageToken=%d&limit=%d",
				url.QueryEscape(string(insertedUser.Email)),
				zerothPage,
				testutil.PageLimit,
			),
			params: api.GetAPIUsersParams{
				Email:     &insertedUser.Email,
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		},
		"get existing user by first name": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{insertedUser, insertedUser2},
			testMsg:      "successfully got users by first name",
			route: fmt.Sprintf(
				"?firstName=%s&pageToken=%d&limit=%d",
				url.QueryEscape(insertedUser.FirstName),
				zerothPage,
				testutil.PageLimit,
			),
			params: api.GetAPIUsersParams{
				Name:      &insertedUser.FirstName,
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		},
		"get existing user by last name": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{insertedUser, insertedUser3},
			testMsg:      "successfully got users by last name",
			route: fmt.Sprintf(
				"?lastName=%s&pageToken=%d&limit=%d",
				url.QueryEscape(insertedUser.LastName),
				zerothPage,
				testutil.PageLimit,
			),
			params: api.GetAPIUsersParams{
				Name:      &insertedUser.LastName,
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		},
		"get users by non-existent query params": {
			httpStatus:   http.StatusOK,
			expectedBody: api.Users{},
			testMsg:      "successfully got empty array of users when users don't exist by query params",
			route: fmt.Sprintf(
				"?lastName=%s&pageToken=%d&limit=%d",
				url.QueryEscape(fakeLastName),
				zerothPage,
				testutil.PageLimit,
			),
			params: api.GetAPIUsersParams{
				Name:      &fakeLastName,
				PageToken: &zerothPage,
				Limit:     testutil.PageLimit,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log("TestUser_GetUsers ranges")
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/users?%s&pageToken=%d&limit=%d", tt.route, zerothPage, testutil.PageLimit),
				nil,
			)
			req.Header.Add("Content-Type", "application/json")

			req.Header.Set(api.ReqHeader, uuid.NewString())

			server.GetAPIUsers(rr, req, tt.params)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respUsers struct {
					Users         api.Users `json:"users"`
					NextPageToken int       `json:"nextPageToken,omitempty"`
				}
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
				require.NoError(t, err, "response body can be decoded into a User")

				require.Equal(t, tt.expectedBody, respUsers.Users, tt.testMsg)
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
	t.Run("route with no query params gets all users", func(t *testing.T) {
		t.Log("TestUser_GetUsers no query all users")
		var tx *sql.Tx
		tx, err = db.Begin()
		require.NoError(t, err, "could not begin transaction")

		var rr *httptest.ResponseRecorder

		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/users?pageToken=%d&limit=%d", zerothPage, testutil.PageLimit),
			nil,
		)
		req.Header.Add("Content-Type", "application/json")

		count := testutil.GetCount(t, db, "User")

		allUsers := api.Users{}
		var pageToken uint32
		for {
			rr = httptest.NewRecorder()
			server.GetAPIUsers(rr, req, api.GetAPIUsersParams{
				PageToken: &pageToken,
				Limit:     testutil.PageLimit,
			},
			)
			require.Equal(t, http.StatusOK, rr.Result().StatusCode)
			var resp struct {
				Users         api.Users `json:"users"`
				NextPageToken int       `json:"nextPageToken"`
			}
			err = json.NewDecoder(rr.Result().Body).Decode(&resp)
			require.NoError(t, err, "response body can be decoded")
			allUsers = append(allUsers, resp.Users...)
			if resp.NextPageToken == 0 {
				break
			}

			pageToken = uint32(resp.NextPageToken)
		}
		err = tx.Commit()
		require.NoError(t, err, "failed to commit transaction")
		testutil.OpenAPIValidateTest(t, rr, req)
		require.Len(t, allUsers, count, "got all users from the User table")
	})
	t.Run("Pagination", func(t *testing.T) {
		t.Log("TestUser_GetUsers pagination")
		for range 11 {
			testutil.InsertUser(t, db)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/users?pageToken=%d&limit=%d", zerothPage, testutil.PageLimit),
			nil,
		)
		req.Header.Set(api.ReqHeader, uuid.NewString())
		params := api.GetAPIUsersParams{
			PageToken: &zerothPage,
			Limit:     testutil.PageLimit,
		}
		server.GetAPIUsers(rr, req, params)
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
		var response struct {
			Users         api.Users `json:"users"`
			NextPageToken int       `json:"nextPageToken,omitempty"`
		}
		err = json.NewDecoder(rr.Result().Body).Decode(&response)
		require.NoError(t, err, "failed to decode first page response")
		require.Len(t, response.Users, 10, "first page should have 10 users")
		if response.NextPageToken != 0 {
			rr2 := httptest.NewRecorder()
			req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users&pageToken=%d", response.NextPageToken), nil)
			req2.Header.Set(api.ReqHeader, uuid.NewString())

			nextToken := uint32(response.NextPageToken)
			params2 := api.GetAPIUsersParams{
				PageToken: &nextToken,
				Limit:     testutil.PageLimit,
			}
			server.GetAPIUsers(rr2, req2, params2)
			require.Equal(t, http.StatusOK, rr2.Result().StatusCode)
			var response2 struct {
				Users         api.Users `json:"users"`
				NextPageToken int       `json:"nextPageToken,omitempty"`
			}
			err = json.NewDecoder(rr2.Result().Body).Decode(&response2)
			require.NoError(t, err, "failed to decode second page response")
			firstPageUserIDs := map[uint32]bool{}
			for _, u := range response.Users {
				firstPageUserIDs[u.Id] = true
			}
			for _, u := range response2.Users {
				require.False(t, firstPageUserIDs[u.Id], "user id %d appears twice", u.Id)
			}
		}
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
