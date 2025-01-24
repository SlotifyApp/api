package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/stretchr/testify/require"

	"github.com/SlotifyApp/slotify-backend/testutil"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestUser_GetUsersUserID(t *testing.T) {
	ctx := context.Background()
	dbh, err := database.NewDatabaseWithContext(ctx)
	defer func() {
		err = dbh.Close()
		require.NoError(t, err, "dbh closed")
	}()
	require.NoError(t, err, "NewDatabaseWithContext doesn't return error")

	var server *api.Server
	server, err = api.NewServerWithContext(ctx, dbh)
	require.NoError(t, err, "NewServerWithContext doesn't return error")

	t.Run("user not found", func(t *testing.T) {
		rr := httptest.NewRecorder()
		userID := 1
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", userID), nil)

		server.GetUsersUserID(rr, req, userID)

		t.Run("it returns 404 when user not found", func(t *testing.T) {
			var errMsg string
			require.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
			require.NoError(t, err, "response body can be decoded into string")
			require.Equal(t, "user doesn't exist", errMsg, "json body has correct message")
		})

		testutil.OpenAPIValidateTest(t, rr, req)
	})

	t.Run("user found", func(t *testing.T) {
		user := api.User{
			Id:        1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@gmail.com",
		}
		// Setup inserting user into db
		var res sql.Result
		query := "INSERT into User (first_name, last_name, email) VALUES(?, ?, ?)"
		res, err = dbh.Exec(query, user.FirstName, user.LastName, user.Email)
		require.NoError(t, err, "dbh insert doesn't return error")
		var rowsAffected int64
		rowsAffected, err = res.RowsAffected()
		require.NoError(t, err, "dbh rowsAffected doesn't return error")
		require.Equal(t, int64(1), rowsAffected, "one user inserted into db")

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.Id), nil)

		server.GetUsersUserID(rr, req, user.Id)

		t.Run("it returns 200 when user is found", func(t *testing.T) {
			var responseUser api.User
			err = json.NewDecoder(rr.Result().Body).Decode(&responseUser)
			require.NoError(t, err, "response body can be decoded into User")

			require.Equal(t, http.StatusOK, rr.Result().StatusCode)
			require.Equal(t, user, responseUser, "response body matches expected user")
		})

		testutil.OpenAPIValidateTest(t, rr, req)
	})
}

func TestUser_PostUsers(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewDatabaseWithContext(ctx)
	defer func() {
		err = db.Close()
		require.NoError(t, err, "dbh closed")
	}()
	require.NoError(t, err, "NewDatabaseWithContext doesn't return error")

	var server *api.Server
	server, err = api.NewServerWithContext(ctx, db)
	require.NoError(t, err, "NewServerWithContext doesn't return error")

	userCreate := api.UserCreate{
		Email:     "sally.doe@gmail.com",
		FirstName: "Sally",
		LastName:  "Doe",
	}
	count := testutil.GetCount(t, db, "User")
	user := api.User{
		Id:        count + 1,
		Email:     userCreate.Email,
		FirstName: userCreate.FirstName,
		LastName:  userCreate.LastName,
	}

	var body []byte
	body, err = json.Marshal(userCreate)
	require.NoError(t, err, "could not marshal json req body user")
	t.Run("new user insert", func(t *testing.T) {
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Add("Content-Type", "application/json")

		server.PostUsers(rr, req)
		// Reset the request body for openapi validate
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		testutil.OpenAPIValidateTest(t, rr, req)

		t.Run("returns 201 on successful insert", func(t *testing.T) {
			var respUser api.User
			require.Equal(t, http.StatusCreated, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&respUser)
			require.NoError(t, err, "response body can be decoded into a User")
			require.Equal(t, user, respUser, "user body correct")
		})
	})

	t.Run("attempt to insert user with email that already exists", func(t *testing.T) {
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Add("Content-Type", "application/json")

		server.PostUsers(rr, req)

		req.Body = io.NopCloser(bytes.NewBuffer(body))
		testutil.OpenAPIValidateTest(t, rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
		var respBody string
		err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
		require.NoError(t, err, "response body can be decoded into a User")
		require.Equal(t, "failed to insert user", respBody, "response body correct")
	})
}

func TestUser_GetUsers(t *testing.T) {
	ctx := context.Background()
	dbh, err := database.NewDatabaseWithContext(ctx)
	defer func() {
		err = dbh.Close()
		require.NoError(t, err, "dbh closed")
	}()
	require.NoError(t, err, "NewDatabaseWithContext doesn't return error")

	var server *api.Server
	server, err = api.NewServerWithContext(ctx, dbh)
	require.NoError(t, err, "NewServerWithContext doesn't return error")

	t.Run("get existing user by email", func(t *testing.T) {
		var email openapi_types.Email = "sally.doe@gmail.com"
		params := api.GetUsersParams{
			Email: &email,
		}
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users?email=%s", *params.Email), nil)
		req.Header.Add("Content-Type", "application/json")

		server.GetUsers(rr, req, params)

		testutil.OpenAPIValidateTest(t, rr, req)

		t.Run("returns 200 with users", func(t *testing.T) {
			var respUsers api.Users
			require.Equal(t, http.StatusOK, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
			require.NoError(t, err, "response body can be decoded into a User")
			require.Len(t, respUsers, 1, "one user returned")
			sally := respUsers[0]
			require.Equal(t, email, sally.Email, "email is correct")
			require.Equal(t, "Sally", sally.FirstName, "first name is correct")
			require.Equal(t, "Doe", sally.LastName, "last name is correct")
		})
	})

	t.Run("get existing user by names", func(t *testing.T) {
		var email openapi_types.Email = "sally.doe@gmail.com"
		firstName := "Sally"
		lastName := "Doe"
		params := api.GetUsersParams{
			FirstName: &firstName,
		}

		t.Run("get existing user by first name", func(t *testing.T) {
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users?firstName=%s", firstName), nil)
			req.Header.Add("Content-Type", "application/json")

			server.GetUsers(rr, req, params)

			testutil.OpenAPIValidateTest(t, rr, req)

			t.Run("returns 200 with users", func(t *testing.T) {
				var respUsers api.Users
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
				require.NoError(t, err, "response body can be decoded into a User")
				require.Len(t, respUsers, 1, "one user returned")
				sally := respUsers[0]
				require.Equal(t, email, sally.Email, "email is correct")
				require.Equal(t, firstName, sally.FirstName, "first name is correct")
				require.Equal(t, lastName, sally.LastName, "last name is correct")
			})
		})

		t.Run("get existing user by last name", func(t *testing.T) {
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users?lastName=%s", lastName), nil)
			req.Header.Add("Content-Type", "application/json")
			params.FirstName = nil
			params.LastName = &lastName

			server.GetUsers(rr, req, params)

			testutil.OpenAPIValidateTest(t, rr, req)

			t.Run("returns 200 with users", func(t *testing.T) {
				var respUsers api.Users
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
				require.NoError(t, err, "response body can be decoded into a User")
				require.Len(t, respUsers, 2, "one user returned")
				john := respUsers[0]
				sally := respUsers[1]
				require.Equal(t, openapi_types.Email("john.doe@gmail.com"), john.Email, "email is correct")
				require.Equal(t, "John", john.FirstName, "first name is correct")
				require.Equal(t, lastName, john.LastName, "last name is correct")

				require.Equal(t, email, sally.Email, "email is correct")
				require.Equal(t, firstName, sally.FirstName, "first name is correct")
				require.Equal(t, lastName, sally.LastName, "last name is correct")
			})
		})
	})

	t.Run("route with no query params gets all users", func(t *testing.T) {
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Add("Content-Type", "application/json")

		server.GetUsers(rr, req, api.GetUsersParams{})

		testutil.OpenAPIValidateTest(t, rr, req)

		var respUsers api.Users
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
		err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
		require.NoError(t, err, "response body can be decoded into a User")
		require.Len(t, respUsers, 2, "all users got")
	})

	t.Run("get non-existing users", func(t *testing.T) {
		// Doesnt exist
		firstName := "blah"
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users?firstName=%s", firstName), nil)
		req.Header.Add("Content-Type", "application/json")

		server.GetUsers(rr, req, api.GetUsersParams{FirstName: &firstName})

		testutil.OpenAPIValidateTest(t, rr, req)

		var respUsers api.Users
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
		err = json.NewDecoder(rr.Result().Body).Decode(&respUsers)
		require.NoError(t, err, "response body can be decoded into a User")
		require.Empty(t, respUsers, "no users with such first name are returned")
	})
}

func TestUser_DeleteUsersUserID(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewDatabaseWithContext(ctx)
	defer func() {
		err = db.Close()
		require.NoError(t, err, "dbh closed")
	}()
	require.NoError(t, err, "NewDatabaseWithContext doesn't return error")

	var server *api.Server
	server, err = api.NewServerWithContext(ctx, db)
	require.NoError(t, err, "NewServerWithContext doesn't return error")

	t.Run("delete user that doesn't exist", func(t *testing.T) {
		rr := httptest.NewRecorder()

		// This user doesnt exist
		userID := 1000
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", userID), nil)
		req.Header.Add("Content-Type", "application/json")

		server.DeleteUsersUserID(rr, req, userID)
		testutil.OpenAPIValidateTest(t, rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)

		var body string
		err = json.NewDecoder(rr.Result().Body).Decode(&body)
		require.NoError(t, err, "decode returns error")
		require.Equal(t, "user not deleted from db", body, "correct response body")
	})

	t.Run("delete user that exists", func(t *testing.T) {
		rr := httptest.NewRecorder()
		userID := 1
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", userID), nil)
		req.Header.Add("Content-Type", "application/json")

		server.DeleteUsersUserID(rr, req, userID)
		testutil.OpenAPIValidateTest(t, rr, req)

		require.Equal(t, http.StatusOK, rr.Result().StatusCode)

		count := testutil.GetCount(t, db, "User")
		require.NoError(t, err, "GetCount returns error")
		require.Equal(t, 1, count, "user deleted from db")

		var body string
		err = json.NewDecoder(rr.Result().Body).Decode(&body)
		require.NoError(t, err, "decode returns error")
		require.Equal(t, "user deleted successfully", body, "user deleted successfully")
	})
}
