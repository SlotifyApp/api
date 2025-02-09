package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

func (s Server) GetAPITeamsJoinableMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	teams, err := s.DB.GetJoinableTeams(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("getting users joinable teams failed: context was cancelled",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("getting users joinable teams: query timed out",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		default:
			s.Logger.Error("failed to get users teams", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError,
				fmt.Sprintf("Sorry, failed to get user id(%d)'s joinable teams", userID))
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, teams)
}

// (GET /api/teams/me).
func (s Server) GetAPITeamsMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
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
func (s Server) GetAPITeams(w http.ResponseWriter, r *http.Request, params GetAPITeamsParams) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	teams, err := s.DB.ListTeams(ctx, params.Name)
	if err != nil {
		s.Logger.Error(zap.Object("params", params), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team api: failed to get teams")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, teams)
}

// (POST /teams) Create a new team and add the user who created the team.
func (s Server) PostAPITeams(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var teamBody PostAPITeamsJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&teamBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Object("body", teamBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

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
			sendError(w, http.StatusBadRequest,
				fmt.Sprintf("team with name %s already exists", teamBody.Name))
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

	notifParams := database.CreateNotificationParams{
		Message: fmt.Sprintf("You successfully created Team %s!", team.Name),
		Created: time.Now(),
	}

	if err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, notifParams); err != nil {
		s.Logger.Errorf("team api: failed to send notification",
			zap.Error(err))
	}

	// TODO: Actually have a db transaction here containing both creating the team
	// or joining the team fails
	// Add user who made the team to the team itself
	s.PostAPITeamsTeamIDUsersMe(w, r, team.Id)
}

// (DELETE /teams/{teamID}) Delete a team by id.
func (s Server) DeleteAPITeamsTeamID(w http.ResponseWriter, r *http.Request, teamID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
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
func (s Server) GetAPITeamsTeamID(w http.ResponseWriter, r *http.Request, teamID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
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
func (s Server) GetAPITeamsTeamIDUsers(w http.ResponseWriter, r *http.Request, teamID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*database.DatabaseTimeout)
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

// (POST /api/teams/{teamID}/users/me).
func (s Server) PostAPITeamsTeamIDUsersMe(w http.ResponseWriter, r *http.Request, teamID uint32) {
	userID, ok := r.Context().Value(UserCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	s.PostAPITeamsTeamIDUsersUserID(w, r, teamID, userID)
}

// (POST /teams/{teamID}/users/{userID}) Add a user to a team.
// nolint: funlen // TODO: Refactor this
func (s Server) PostAPITeamsTeamIDUsersUserID(w http.ResponseWriter, r *http.Request, teamID uint32, userID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*database.DatabaseTimeout)
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

	t, err := s.DB.GetTeamByID(ctx, teamID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("team api: add user to team: get team by id: context cancelled",
				zap.Uint32("userID", userID), zap.Uint32("teamID", teamID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "team api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("team api: add user to team: get team by id query timed out",
				zap.Uint32("userID", userID), zap.Uint32("teamID", teamID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "team api: add team query timed out")
			return
		default:
			s.Logger.Errorf("team api: add user to team: get team by id query timed out",
				zap.Error(err))
			sendError(w, http.StatusInternalServerError,
				"team api: added team but failed to get team by id")
			return
		}
	}

	dbParams := database.GetAllTeamMembersExceptParams{
		TeamID: teamID,
		UserID: userID,
	}
	members, err := s.DB.GetAllTeamMembersExcept(ctx, dbParams)
	if err != nil {
		s.Logger.Errorf("team api: failed to get team members except new for PostAPITeamsTeamIDUsersUserID",
			zap.Uint32("exceptUserID", userID),
			zap.Error(err))
		// Still return ok because the actual endpoint worked, its just notifications that didnt
		SetHeaderAndWriteResponse(w, http.StatusOK, t)
		return
	}

	u, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		s.Logger.Errorf("team api: failed to get user details for PostAPITeamsTeamIDUsersUserID",
			zap.Error(err))
		// Still return ok because the actual endpoint worked, its just notifications that didnt
		SetHeaderAndWriteResponse(w, http.StatusCreated, t)
		return
	}
	allMemberNotif := database.CreateNotificationParams{
		Message: fmt.Sprintf("Say hi to %s, he just joined Team %s", u.FirstName+" "+u.LastName, t.Name),
		Created: time.Now(),
	}

	newMemberNotif := database.CreateNotificationParams{
		Message: fmt.Sprintf("You were added to Team %s!", t.Name),
		Created: time.Now(),
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, members, allMemberNotif)
	if err != nil {
		// Don't return error, attempt to send individual notification too
		s.Logger.Errorf(
			"team api: failed to send notification to all existing users of team, adding team member",
			zap.Error(err))
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, newMemberNotif)
	if err != nil {
		s.Logger.Errorf("team api: failed to send notification PostAPITeamsTeamIDUsersUserID to user that just joined team",
			zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, t)
}

func (s Server) OptionsAPITeams(w http.ResponseWriter, _ *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")   // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}
