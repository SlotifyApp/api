package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/jwt"
	"go.uber.org/zap"
)

// (GET /api/teams/me).
func (s Server) GetAPITeamsMe(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.GetUserIDFromReq(r)
	if err != nil {
		s.Logger.Error("failed to get userid from request access token")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()
	teams, err := s.DB.GetUsersTeams(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("getting users team failed: context was cancelled",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("getting users team query timed out",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		default:
			s.Logger.Error("failed to get users teams", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Sorry, failed to get user id(%d)'s teams", userID))
			return
		}
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, teams)
}

// (GET /teams) Get a team by query params.
func (s Server) GetAPITeams(w http.ResponseWriter, _ *http.Request, params GetAPITeamsParams) {
	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()

	teams, err := s.DB.ListTeams(ctx, params.Name)
	if err != nil {
		s.Logger.Error(zap.Object("params", params), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team api: failed to get teams")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, teams)
}

// (POST /teams) Create a new team.
func (s Server) PostAPITeams(w http.ResponseWriter, r *http.Request) {
	var teamBody PostAPITeamsJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&teamBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Object("body", teamBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()
	var teamID int64
	teamID, err = s.DB.AddTeam(ctx, teamBody.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: add team query context cancelled",
				zap.Object("req_body", teamBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "something went wrong")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: add team query timed out",
				zap.Object("req_body", teamBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "add team query timed out")
			return
		case database.IsDuplicateEntrySQLError(err):
			s.Logger.Error("team api: team already exists", zap.Object("req_body", teamBody), zap.Error(err))
			sendError(w, http.StatusBadRequest, fmt.Sprintf("team with name %s already exists", teamBody.Name))
			return
		default:
			s.Logger.Error("failed to create team", zap.Object("body", teamBody), zap.Error(err))
			sendError(w, http.StatusBadRequest, "team api: team creation unsuccessful")
			return
		}
	}

	team := Team{
		//nolint: gosec // id is unsigned 32 bit int
		Id:   uint32(teamID),
		Name: teamBody.Name,
	}
	SetHeaderAndWriteResponse(w, http.StatusCreated, team)
}

// (DELETE /teams/{teamID}) Delete a team by id.
func (s Server) DeleteAPITeamsTeamID(w http.ResponseWriter, _ *http.Request, teamID uint32) {
	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()

	rowsDeleted, err := s.DB.DeleteTeamByID(ctx, teamID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: delete team by id query timed out", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "add team query timed out")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: delete team by id query timed out", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "add team query timed out")
			return
		default:
			s.Logger.Error("team api: failed to DeleteTeamsTeamID", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "team api: team deletion unsuccessful")
			return
		}
	}

	if rowsDeleted != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsDeleted,
			ExpectedRows: []int64{1},
		}
		s.Logger.Errorf("team api failed to delete team", zap.Error(err))
		sendError(w, http.StatusBadRequest, "team api: incorrect team id")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "team api: team deleted successfully")
}

// (GET /teams/{teamID}) Get a team by id.
func (s Server) GetAPITeamsTeamID(w http.ResponseWriter, _ *http.Request, teamID uint32) {
	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()

	team, err := s.DB.GetTeamByID(ctx, teamID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: context cancelled", zap.Uint32("teamID", teamID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "sorry try again")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: delete team by id query timed out", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "add team query timed out")
			return
		case errors.Is(err, sql.ErrNoRows):
			s.Logger.Error("team api: failed to get team by id, no matching rows", zap.Error(err))
			sendError(w, http.StatusNotFound, fmt.Sprintf("team api: team with id %d does not exist", teamID))
			return
		default:
			s.Logger.Error("team api: failed to GetTeamsTeamID", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "team api: failed to get team")
			return
		}
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, team)
}

// (GET /teams/{teamID}/users) Get all members of a team.
func (s Server) GetAPITeamsTeamIDUsers(w http.ResponseWriter, _ *http.Request, teamID uint32) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*database.DatabaseTimeout)
	defer cancel()

	count, err := s.DB.CountTeamByID(ctx, teamID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: context cancelled", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "sorry try again")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: get all team members query timed out", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "get all team members query timed out")
			return
		default:
			s.Logger.Error("team api: failed to count team members by id", zap.Uint32("teamID", teamID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "get all team members query timed out")
			return
		}
	}

	if count == 0 {
		s.Logger.Error("team api: team members of non-existent team requested", zap.Uint32("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusForbidden, fmt.Sprintf("team api: team with id(%d) does not exist", teamID))
		return
	}

	users, err := s.DB.GetAllTeamMembers(ctx, teamID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: context cancelled", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: get all team members query timed out", zap.Uint32("teamID", teamID))
			sendError(w, http.StatusInternalServerError, "get all team members query timed out")
			return
		default:
			s.Logger.Errorf("team api: failed to get all team members: %w", err)
			sendError(w, http.StatusInternalServerError,
				fmt.Sprintf("team api: failed to get team members for team with id %d", teamID))
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}

// (POST /teams/{teamID}/users/{userID}) Add a user to a team.
func (s Server) PostAPITeamsTeamIDUsersUserID(w http.ResponseWriter, _ *http.Request, teamID uint32, userID uint32) {
	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()

	rowsAffected, err := s.DB.AddUserToTeam(ctx, database.AddUserToTeamParams{
		UserID: userID,
		TeamID: teamID,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: context cancelled",
				zap.Uint32("userID", userID), zap.Uint32("teamID", teamID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "team api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: add user to team query timed out",
				zap.Uint32("userID", userID), zap.Uint32("teamID", teamID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "team api: add team query timed out")
			return
		case database.IsRowDoesNotExistSQLError(err):
			s.Logger.Errorf("team api: user id or team id invalid fk: %w", err)
			sendError(w, http.StatusForbidden,
				fmt.Sprintf("team api: team id(%d) or user id(%d) was invalid", teamID, userID))
			return
		default:
			s.Logger.Errorf("team api: failed to add user to team: %w", err)
			sendError(w, http.StatusInternalServerError, "team api: failed to add user to team")
			return
		}
	}

	if rowsAffected != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{1},
		}
		s.Logger.Error(zap.Uint32("userID", userID), zap.Uint32("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "user api: failed to add member to team")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK,
		fmt.Sprintf("team api: member with id %d added to team with id %d", userID, teamID))
}
