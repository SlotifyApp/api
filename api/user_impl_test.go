package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saths008/slotify-backend/api"
	"github.com/saths008/slotify-backend/database"
	"github.com/stretchr/testify/require"
	validator "openapi.tanna.dev/go/validator/openapi3"
)

// Test helper to ensure request matches OpenAPI spec.
func OpenAPIValidateTestHelper(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
	t.Run("it matches OpenAPI", func(t *testing.T) {
		doc, err := api.GetSwagger()
		require.NoError(t, err, "GetSwagger doesn't return error")

		_ = validator.NewValidator(doc).ForTest(t, rr, req)
	})
}

func TestServer_GetUsersUserID(t *testing.T) {
	ctx := context.Background()
	dbh, err := database.NewDatabaseWithContext(ctx)
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
			require.Equal(t, 404, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
			require.NoError(t, err, "response body can be decoded into string")
			require.Equal(t, "user doesn't exist", errMsg, "json body has correct message")
		})

		OpenAPIValidateTestHelper(t, rr, req)
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

			require.Equal(t, 200, rr.Result().StatusCode)
			require.Equal(t, user, responseUser, "response body matches expected user")
		})

		OpenAPIValidateTestHelper(t, rr, req)
	})
}
