package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

// (GET /teams) Get a team by query params.
func (s Server) GetTeams(w http.ResponseWriter, _ *http.Request, params GetTeamsParams) {
	teams, err := s.TeamRepository.GetTeamsByQueryParams(params)
	if err != nil {
		s.Logger.Error(zap.Object("params", params), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team api: failed to get teams")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, teams)
}

// (POST /teams) Create a new team.
func (s Server) PostTeams(w http.ResponseWriter, r *http.Request) {
	var teamBody PostTeamsJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&teamBody); err != nil {
		errMsg := "failed to unmarshal request body correctly"
		s.Logger.Error(errMsg, zap.Object("body", teamBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}
	var team Team
	if team, err = s.TeamRepository.AddTeam(teamBody); err != nil {
		if database.IsDuplicateEntrySQLError(err) {
			s.Logger.Error("team api: team already exists", zap.Object("req_body", teamBody), zap.Error(err))
			sendError(w, http.StatusBadRequest, fmt.Sprintf("team with name %s already exists", teamBody.Name))
			return
		}
		s.Logger.Error("failed to create team", zap.Object("body", teamBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team api: team creation unsuccessful")
		return
	}
	SetHeaderAndWriteResponse(w, http.StatusCreated, team)
}

// (DELETE /teams/{teamID}) Delete a team by id.
func (s Server) DeleteTeamsTeamID(w http.ResponseWriter, _ *http.Request, teamID int) {
	if err := s.TeamRepository.DeleteTeamByID(teamID); err != nil {
		if errors.Is(err, database.WrongNumberSQLRowsError{}) {
			s.Logger.Error("team api: incorrect team ID", zap.Error(err))
			sendError(w, http.StatusBadRequest, "team api: incorrect team ID")
			return
		}
		s.Logger.Error("team api: failed to DeleteTeamsTeamID", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "team api: team deletion unsuccessful")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "team api: team deleted successfully")
}

// (GET /teams/{teamID}) Get a team by id.
func (s Server) GetTeamsTeamID(w http.ResponseWriter, _ *http.Request, teamID int) {
	team, err := s.TeamRepository.GetTeamByID(teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Error("team api: failed to GetTeamsTeamID, no matching rows", zap.Error(err))
			sendError(w, http.StatusNotFound, fmt.Sprintf("team api: team with id %d does not exist", teamID))
			return
		}

		s.Logger.Error("team api: failed to GetTeamsTeamID", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "team api: failed to get team")
		return
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, team)
}

// (GET /teams/{teamID}/users) Get all members of a team.
func (s Server) GetTeamsTeamIDUsers(w http.ResponseWriter, _ *http.Request, teamID int) {
	var users Users
	var err error
	if users, err = s.TeamRepository.GetAllTeamMembers(teamID); err != nil {
		if errors.Is(err, database.ErrTeamIDInvalid) {
			s.Logger.Errorf("team api: team id was invalid: %w", err)
			sendError(w, http.StatusForbidden,
				fmt.Sprintf("team api: team with id(%d) does not exist", teamID))
			return
		}
		s.Logger.Errorf("team api: failed to GetTeamsTeamIDUsers: %w", err)
		sendError(w, http.StatusInternalServerError,
			fmt.Sprintf("team api: failed to get team members for team with id %d", teamID))
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}

// (POST /teams/{teamID}/users/{userID}) Add a user to a team.
func (s Server) PostTeamsTeamIDUsersUserID(w http.ResponseWriter, _ *http.Request, teamID int, userID int) {
	if err := s.TeamRepository.AddUserToTeam(teamID, userID); err != nil {
		if database.IsRowDoesNotExistSQLError(err) {
			s.Logger.Errorf("team api: user id or team id invalid fk: %w", err)
			sendError(w, http.StatusForbidden,
				fmt.Sprintf("team api: team id(%d) or user id(%d) was invalid", teamID, userID))
			return
		}
		s.Logger.Errorf("team api: failed to add user to team: %w", err)
		sendError(w, http.StatusInternalServerError, "team api: failed to add user to team")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK,
		fmt.Sprintf("team api: member with id %d added to team with id %d", userID, teamID))
}
