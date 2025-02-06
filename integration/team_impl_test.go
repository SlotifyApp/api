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
	"github.com/SlotifyApp/slotify-backend/testutil"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestTeam_GetTeams(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, context.Background())
	// For testing, we want the underlying db connection rather than the
	// sqlc queries.
	db := database.DB

	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	insertedTeam := testutil.InsertTeam(t, db)

	tests := map[string]struct {
		httpStatus    int
		teamName      string
		expectedTeams api.Teams
		testMsg       string
		route         string
	}{
		"team does not exist": {
			httpStatus:    http.StatusOK,
			teamName:      "DoesntExist",
			expectedTeams: api.Teams{},
			testMsg:       "empty array is returned when team does not exist",
		},
		"team does exist": {
			httpStatus:    http.StatusOK,
			teamName:      insertedTeam.Name,
			expectedTeams: api.Teams{insertedTeam},
			testMsg:       "correct array is returned when team exists",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			params := api.GetAPITeamsParams{
				Name: &tt.teamName,
			}
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/api/teams?name=%s", url.QueryEscape(*params.Name)), nil)

			server.GetAPITeams(rr, req, params)

			var teams api.Teams
			require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&teams)
			require.NoError(t, err, "response cannot be decoded into Teams struct")
			require.Equal(t, tt.expectedTeams, teams, tt.testMsg)

			testutil.OpenAPIValidateTest(t, rr, req)
		})
	}
}

func TestTeam_PostTeams(t *testing.T) {
	t.Parallel()

	database, server := testutil.NewServerAndDB(t, context.Background())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})
	user := testutil.InsertUser(t, db)
	// Setup
	insertedTeam := testutil.InsertTeam(t, db)
	newTeamName := gofakeit.ProductName()

	tests := map[string]struct {
		httpStatus       int
		teamBody         api.TeamCreate
		teamName         string
		expectedRespBody any
		testMsg          string
	}{
		"team inserted correctly": {
			httpStatus: http.StatusCreated,
			teamName:   newTeamName,
			teamBody: api.TeamCreate{
				Name: newTeamName,
			},
			testMsg: "team made successfully",
		},
		"team already exists": {
			httpStatus: http.StatusBadRequest,
			teamBody: api.TeamCreate{
				Name: insertedTeam.Name,
			},
			expectedRespBody: fmt.Sprintf("team with name %s already exists", insertedTeam.Name),
			testMsg:          "team that already exists",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.teamBody)
			require.NoError(t, err, "could not marshal json req body team")
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/teams", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), "userID", user.Id)
			req = req.WithContext(ctx)

			server.PostAPITeams(rr, req)

			if tt.httpStatus == http.StatusCreated {
				var team api.Team
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&team)
				require.NoError(t, err, "response cannot be decoded into Team struct")

				require.Equal(t, tt.teamName, team.Name, tt.testMsg)
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

func TestTeam_DeleteTeamsTeamID(t *testing.T) {
	t.Parallel()

	var err error

	database, server := testutil.NewServerAndDB(t, context.Background())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	teamInserted := testutil.InsertTeam(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           uint32
		testMsg          string
	}{
		"deleting team that doesn't exist": {
			expectedRespBody: "team api: incorrect team id",
			httpStatus:       http.StatusBadRequest,
			teamID:           10000,
			testMsg:          "team that doesn't exist, returns client error",
		}, "deleting team that exists": {
			expectedRespBody: "team api: team deleted successfully",
			httpStatus:       http.StatusOK,
			teamID:           teamInserted.Id,
			testMsg:          "deleting team that exists is successful",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/teams/%d", tt.teamID), nil)

			server.DeleteAPITeamsTeamID(rr, req, tt.teamID)

			testutil.OpenAPIValidateTest(t, rr, req)
			var errMsg string
			require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&errMsg)
			require.NoError(t, err, "response cannot be decoded into string")
			require.Equal(t, tt.expectedRespBody, errMsg, tt.testMsg)
		})
	}
}

func TestTeam_GetTeamsTeamID(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, context.Background())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	teamInserted := testutil.InsertTeam(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           uint32
		testMsg          string
	}{
		"get team that exists": {
			expectedRespBody: teamInserted,
			httpStatus:       http.StatusOK,
			teamID:           teamInserted.Id,
			testMsg:          "can get team that exists successfully",
		},
		"get team that doesn't exist": {
			expectedRespBody: "team api: team with id 1000 does not exist",
			httpStatus:       http.StatusNotFound,
			teamID:           1000,
			testMsg:          "deleting team that doesn't exist is unsuccessful",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/teams/%d", tt.teamID), nil)

			server.GetAPITeamsTeamID(rr, req, tt.teamID)

			testutil.OpenAPIValidateTest(t, rr, req)
			if tt.httpStatus == http.StatusOK {
				var team api.Team
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&team)
				require.NoError(t, err, "response cannot be decoded into a team")
				require.Equal(t, tt.expectedRespBody, team, tt.testMsg)
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

func TestTeam_PostTeamsTeamIDUsersUserID(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, context.Background())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	teamInserted := testutil.InsertTeam(t, db)

	userInserted := testutil.InsertUser(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           uint32
		userID           uint32
		testMsg          string
	}{
		"insert a user into a team": {
			expectedRespBody: api.Team{
				Id:   teamInserted.Id,
				Name: teamInserted.Name,
			},
			httpStatus: http.StatusCreated,
			userID:     userInserted.Id,
			teamID:     teamInserted.Id,
			testMsg:    "can add user to team where both exist successfully",
		},
		"insert a user into a non-existent team": {
			expectedRespBody: fmt.Sprintf("team api: team id(%d) or user id(%d) was invalid",
				1000, userInserted.Id),
			httpStatus: http.StatusForbidden,
			userID:     userInserted.Id,
			teamID:     1000,
			testMsg:    "cannot add user to team where team does not exist",
		},

		"insert an non-existent user into a team": {
			expectedRespBody: fmt.Sprintf("team api: team id(%d) or user id(%d) was invalid", teamInserted.Id, 10000),
			httpStatus:       http.StatusForbidden,
			userID:           10000,
			teamID:           teamInserted.Id,
			testMsg:          "cannot add user to team where user does not exist",
		},

		"user and team ids do not exist": {
			expectedRespBody: fmt.Sprintf("team api: team id(%d) or user id(%d) was invalid", 10000, 10000),
			httpStatus:       http.StatusForbidden,
			userID:           10000,
			teamID:           10000,
			testMsg:          "cannot add user to team where neither exists",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/teams/%d/users/%d", tt.teamID, tt.userID), nil)

			server.PostAPITeamsTeamIDUsersUserID(rr, req, tt.teamID, tt.userID)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusCreated {
				var team api.Team
				require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
				err = json.NewDecoder(rr.Result().Body).Decode(&team)
				require.NoError(t, err, "response cannot be decoded into Team struct")
				require.Equal(t, tt.expectedRespBody, team, tt.testMsg)
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

func TestTeam_GetTeamsTeamIDUsers(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, context.Background())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	teamInserted := testutil.InsertTeam(t, db)

	userInserted := testutil.InsertUser(t, db)
	userInserted2 := testutil.InsertUser(t, db)

	testutil.AddUserToTeam(t, db, userInserted.Id, teamInserted.Id)
	testutil.AddUserToTeam(t, db, userInserted2.Id, teamInserted.Id)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           uint32
		testMsg          string
	}{
		"get members of a non-existing team": {
			expectedRespBody: "team api: team with id(10000) does not exist",
			httpStatus:       http.StatusForbidden,
			teamID:           10000,
			testMsg:          "correct error returns when team doesn't exist",
		},
		"get members of an existing team": {
			expectedRespBody: api.Users{userInserted, userInserted2},
			httpStatus:       http.StatusOK,
			teamID:           teamInserted.Id,
			testMsg:          "correctly get members of a team",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/teams/%d/users", tt.teamID), nil)

			server.GetAPITeamsTeamIDUsers(rr, req, tt.teamID)

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

func TestTeam_GetAPITeamsMe(t *testing.T) {
	t.Parallel()

	var err error
	database, server := testutil.NewServerAndDB(t, context.Background())
	db := database.DB
	t.Cleanup(func() {
		testutil.CloseDB(db)
	})

	user1 := testutil.InsertUser(t, db, testutil.WithEmail("blah@example.com"))
	jwt1 := testutil.CreateJWT(t, user1.Id, user1.Email)

	insertedTeam1 := testutil.InsertTeam(t, db)
	insertedTeam2 := testutil.InsertTeam(t, db)
	user2 := testutil.InsertUser(t, db)
	testutil.AddUserToTeam(t, db, user2.Id, insertedTeam1.Id)
	testutil.AddUserToTeam(t, db, user2.Id, insertedTeam2.Id)
	jwt2 := testutil.CreateJWT(t, user2.Id, user2.Email)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		jwt              string
		testMsg          string
	}{
		"get teams of user who has no teams": {
			expectedRespBody: api.Teams{},
			httpStatus:       http.StatusOK,
			jwt:              jwt1,
			testMsg:          "user who has no teams returns empty list",
		},
		"get teams of user who has many teams": {
			expectedRespBody: api.Teams{insertedTeam1, insertedTeam2},
			httpStatus:       http.StatusOK,
			jwt:              jwt2,
			testMsg:          "correctly get all of a user's teams",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/teams/me", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.jwt))

			server.GetAPITeamsMe(rr, req)

			testutil.OpenAPIValidateTest(t, rr, req)

			if tt.httpStatus == http.StatusOK {
				var respBody api.Teams
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
