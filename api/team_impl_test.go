package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/stretchr/testify/require"
)

func TestTeam_GetTeams(t *testing.T) {
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

	t.Run("test with team that doesn't exist", func(t *testing.T) {
		teamName := "Doesn'tExistTeam"
		params := api.GetTeamsParams{
			Name: &teamName,
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/teams?name=%s", teamName), nil)

		server.GetTeams(rr, req, params)

		var teams api.Teams
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
		err = json.NewDecoder(rr.Result().Body).Decode(&teams)
		require.NoError(t, err, "response body can be decoded into string")
		require.Empty(t, teams, "no teams are returned")

		testutil.OpenAPIValidateTest(t, rr, req)
	})

	// t.Run("user found", func(t *testing.T) {
	// 	user := api.User{
	// 		Id:        1,
	// 		FirstName: "John",
	// 		LastName:  "Doe",
	// 		Email:     "john.doe@gmail.com",
	// 	}
	// 	// Setup inserting user into db
	// 	var res sql.Result
	// 	query := "INSERT into User (first_name, last_name, email) VALUES(?, ?, ?)"
	// 	res, err = dbh.Exec(query, user.FirstName, user.LastName, user.Email)
	// 	require.NoError(t, err, "dbh insert doesn't return error")
	// 	var rowsAffected int64
	// 	rowsAffected, err = res.RowsAffected()
	// 	require.NoError(t, err, "dbh rowsAffected doesn't return error")
	// 	require.Equal(t, int64(1), rowsAffected, "one user inserted into db")
	//
	// 	rr := httptest.NewRecorder()
	// 	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.Id), nil)
	//
	// 	server.GetUsersUserID(rr, req, user.Id)
	//
	// 	t.Run("it returns 200 when user is found", func(t *testing.T) {
	// 		var responseUser api.User
	// 		err = json.NewDecoder(rr.Result().Body).Decode(&responseUser)
	// 		require.NoError(t, err, "response body can be decoded into User")
	//
	// 		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
	// 		require.Equal(t, user, responseUser, "response body matches expected user")
	// 	})
	//
	// 	testutil.OpenAPIValidateTestHelper(t, rr, req)
	// })
}
