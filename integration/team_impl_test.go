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
	"github.com/stretchr/testify/require"
)

func TestTeam_GetTeams(t *testing.T) {
	var err error
	db, server := testutil.NewServerAndDB(t, context.Background())
	defer testutil.CloseDB(db)

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
			params := api.GetTeamsParams{
				Name: &tt.teamName,
			}
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/teams?name=%s", url.QueryEscape(*params.Name)), nil)

			server.GetTeams(rr, req, params)

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
	db, server := testutil.NewServerAndDB(t, context.Background())
	defer testutil.CloseDB(db)

	insertedTeam := testutil.InsertTeam(t, db)

	tests := map[string]struct {
		httpStatus       int
		teamBody         api.TeamCreate
		teamName         string
		expectedRespBody any
		testMsg          string
	}{
		"team inserted correctly": {
			httpStatus: 201,
			teamName:   "NewTeam",
			teamBody: api.TeamCreate{
				Name: "NewTeam",
			},
			expectedRespBody: api.Team{
				Id:   testutil.GetNextAutoIncrementValue(t, db, "Team") - 1,
				Name: "NewTeam",
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
			body, err := json.Marshal(tt.teamBody)
			require.NoError(t, err, "could not marshal json req body team")
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/teams", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")

			server.PostTeams(rr, req)

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
	var err error
	db, server := testutil.NewServerAndDB(t, context.Background())
	defer testutil.CloseDB(db)

	teamInserted := testutil.InsertTeam(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           int
		testMsg          string
	}{
		"deleting team that doesn't exist": {
			expectedRespBody: "team api: incorrect team ID",
			httpStatus:       http.StatusBadRequest,
			teamID:           1000,
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
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/teams/%d", tt.teamID), nil)

			server.DeleteTeamsTeamID(rr, req, tt.teamID)

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
	var err error
	db, server := testutil.NewServerAndDB(t, context.Background())
	defer testutil.CloseDB(db)

	teamInserted := testutil.InsertTeam(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           int
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
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/teams/%d", tt.teamID), nil)

			server.GetTeamsTeamID(rr, req, tt.teamID)

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
	var err error
	db, server := testutil.NewServerAndDB(t, context.Background())
	defer testutil.CloseDB(db)

	teamInserted := testutil.InsertTeam(t, db)

	userInserted := testutil.InsertUser(t, db)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           int
		userID           int
		testMsg          string
	}{
		"insert an existing user into an existing team": {
			expectedRespBody: fmt.Sprintf("team api: member with id %d added to team with id %d",
				userInserted.Id, teamInserted.Id),
			httpStatus: http.StatusOK,
			userID:     userInserted.Id,
			teamID:     teamInserted.Id,
			testMsg:    "can add user to team where both exist successfully",
		},
		"insert an existing user into a non-existing team": {
			expectedRespBody: fmt.Sprintf("team api: team id(%d) or user id(%d) was invalid",
				1000, userInserted.Id),
			httpStatus: http.StatusForbidden,
			userID:     userInserted.Id,
			teamID:     1000,
			testMsg:    "cannot add user to team where team does not exist",
		},

		"insert an non-existing user into a existing team": {
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
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/teams/%d/users/%d", tt.teamID, tt.userID), nil)

			server.PostTeamsTeamIDUsersUserID(rr, req, tt.teamID, tt.userID)

			testutil.OpenAPIValidateTest(t, rr, req)

			var respBody string
			require.Equal(t, tt.httpStatus, rr.Result().StatusCode)
			err = json.NewDecoder(rr.Result().Body).Decode(&respBody)
			require.NoError(t, err, "response cannot be decoded into string")
			require.Equal(t, tt.expectedRespBody, respBody, tt.testMsg)
		})
	}
}

func TestTeam_GetTeamsTeamIDUsers(t *testing.T) {
	var err error
	db, server := testutil.NewServerAndDB(t, context.Background())
	defer testutil.CloseDB(db)

	teamInserted := testutil.InsertTeam(t, db)

	userInserted := testutil.InsertUser(t, db)
	userInserted2 := testutil.InsertUser(t, db)

	testutil.AddUserToTeam(t, db, userInserted.Id, teamInserted.Id)
	testutil.AddUserToTeam(t, db, userInserted2.Id, teamInserted.Id)

	tests := map[string]struct {
		expectedRespBody any
		httpStatus       int
		teamID           int
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
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/teams/%d/users", tt.teamID), nil)

			server.GetTeamsTeamIDUsers(rr, req, tt.teamID)

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
