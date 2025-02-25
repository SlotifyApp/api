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

// (GET /api/slotify-groups/joinable/me).
func (s Server) GetAPISlotifyGroupsJoinableMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	slotifyGroups, err := s.DB.GetJoinableSlotifyGroups(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("getting users joinable slotifyGroups failed: context was cancelled",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("getting users joinable slotifyGroups: query timed out",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		default:
			s.Logger.Error("failed to get users slotifyGroups", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError,
				fmt.Sprintf("Sorry, failed to get user id(%d)'s joinable slotifyGroups", userID))
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, slotifyGroups)
}

// (GET /api/slotify-groups/me).
func (s Server) GetAPISlotifyGroupsMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	slotifyGroups, err := s.DB.GetUsersSlotifyGroups(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("getting users slotifyGroup failed: context was cancelled",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("getting users slotifyGroup query timed out",
				zap.Uint32("userID", userID))
			sendError(w, http.StatusUnauthorized, "Try again later.")
			return
		default:
			s.Logger.Error("failed to get users slotifyGroups", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Sorry, failed to get user id(%d)'s slotifyGroups", userID))
			return
		}
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, slotifyGroups)
}

// (GET /api/slotify-groups).
func (s Server) GetAPISlotifyGroups(w http.ResponseWriter, r *http.Request, params GetAPISlotifyGroupsParams) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	slotifyGroups, err := s.DB.ListSlotifyGroups(ctx, params.Name)
	if err != nil {
		s.Logger.Error(zap.Object("params", params), zap.Error(err))
		sendError(w, http.StatusBadRequest, "slotifyGroup api: failed to get slotifyGroups")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, slotifyGroups)
}

// (POST /api/slotify-groups).
func (s Server) PostAPISlotifyGroups(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var slotifyGroupBody PostAPISlotifyGroupsJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&slotifyGroupBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Object("body", slotifyGroupBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	var slotifyGroupID int64
	slotifyGroupID, err = s.DB.AddSlotifyGroup(ctx, slotifyGroupBody.Name)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: add slotifyGroup query context cancelled",
				zap.Object("req_body", slotifyGroupBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "something went wrong")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: add slotifyGroup query timed out",
				zap.Object("req_body", slotifyGroupBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "add slotifyGroup query timed out")
			return
		case database.IsDuplicateEntrySQLError(err):
			s.Logger.Error("slotifyGroup api: slotifyGroup already exists",
				zap.Object("req_body", slotifyGroupBody), zap.Error(err))
			sendError(w, http.StatusBadRequest,
				fmt.Sprintf("slotifyGroup with name %s already exists", slotifyGroupBody.Name))
			return
		default:
			s.Logger.Error("failed to create slotifyGroup", zap.Object("body", slotifyGroupBody), zap.Error(err))
			sendError(w, http.StatusBadRequest, "slotifyGroup api: slotifyGroup creation unsuccessful")
			return
		}
	}

	slotifyGroup := SlotifyGroup{
		//nolint: gosec // id is unsigned 32 bit int
		Id:   uint32(slotifyGroupID),
		Name: slotifyGroupBody.Name,
	}

	notifParams := database.CreateNotificationParams{
		Message: fmt.Sprintf("You successfully created SlotifyGroup %s!", slotifyGroup.Name),
		Created: time.Now(),
	}

	if err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, notifParams); err != nil {
		s.Logger.Errorf("slotifyGroup api: failed to send notification",
			zap.Error(err))
	}

	// TODO: Actually have a db transaction here containing both creating the slotifyGroup
	// or joining the slotifyGroup fails
	// Add user who made the slotifyGroup to the slotifyGroup itself
	s.PostAPISlotifyGroupsSlotifyGroupIDUsersMe(w, r, slotifyGroup.Id)
}

// (DELETE /api/slotify-groups/{slotifyGroupID}).
func (s Server) DeleteAPISlotifyGroupsSlotifyGroupID(w http.ResponseWriter, r *http.Request, slotifyGroupID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	rowsDeleted, err := s.DB.DeleteSlotifyGroupByID(ctx, slotifyGroupID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: delete slotifyGroup by id query timed out",
				zap.Uint32("slotifyGroupID", slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "add slotifyGroup query timed out")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: delete slotifyGroup by id query timed out",
				zap.Uint32("slotifyGroupID", slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "add slotifyGroup query timed out")
			return
		default:
			s.Logger.Error("slotifyGroup api: failed to DeleteSlotifyGroupsSlotifyGroupID", zap.Error(err))
			sendError(w, http.StatusInternalServerError,
				"slotifyGroup api: slotifyGroup deletion unsuccessful")
			return
		}
	}

	if rowsDeleted != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsDeleted,
			ExpectedRows: []int64{1},
		}
		s.Logger.Errorf("slotifyGroup api failed to delete slotifyGroup", zap.Error(err))
		sendError(w, http.StatusBadRequest, "slotifyGroup api: incorrect slotifyGroup id")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "slotifyGroup api: slotifyGroup deleted successfully")
}

// (GET /api/slotify-groups/{slotifyGroupID}).
func (s Server) GetAPISlotifyGroupsSlotifyGroupID(w http.ResponseWriter, r *http.Request, slotifyGroupID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	slotifyGroup, err := s.DB.GetSlotifyGroupByID(ctx, slotifyGroupID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: context cancelled",
				zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "sorry try again")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: delete slotifyGroup by id query timed out",
				zap.Uint32("slotifyGroupID", slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "add slotifyGroup query timed out")
			return
		case errors.Is(err, sql.ErrNoRows):
			s.Logger.Error("slotifyGroup api: failed to get slotifyGroup by id, no matching rows",
				zap.Error(err))
			sendError(w, http.StatusNotFound,
				fmt.Sprintf("slotifyGroup api: slotifyGroup with id %d does not exist", slotifyGroupID))
			return
		default:
			s.Logger.Error("slotifyGroup api: failed to GetSlotifyGroupsSlotifyGroupID", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "slotifyGroup api: failed to get slotifyGroup")
			return
		}
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, slotifyGroup)
}

// (GET /api/slotify-groups/{slotifyGroupID}/users).
func (s Server) GetAPISlotifyGroupsSlotifyGroupIDUsers(w http.ResponseWriter, r *http.Request, slotifyGroupID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*database.DatabaseTimeout)
	defer cancel()

	count, err := s.DB.CountSlotifyGroupByID(ctx, slotifyGroupID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: context cancelled", zap.Uint32("slotifyGroupID",
				slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "sorry try again")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: get all slotifyGroup members query timed out",
				zap.Uint32("slotifyGroupID", slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "get all slotifyGroup members query timed out")
			return
		default:
			s.Logger.Error("slotifyGroup api: failed to count slotifyGroup members by id",
				zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "get all slotifyGroup members query timed out")
			return
		}
	}

	if count == 0 {
		s.Logger.Error("slotifyGroup api: slotifyGroup members of non-existent slotifyGroup requested",
			zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
		sendError(w, http.StatusForbidden,
			fmt.Sprintf("slotifyGroup api: slotifyGroup with id(%d) does not exist", slotifyGroupID))
		return
	}

	users, err := s.DB.GetAllSlotifyGroupMembers(ctx, slotifyGroupID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: context cancelled", zap.Uint32("slotifyGroupID", slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: get all slotifyGroup members query timed out",
				zap.Uint32("slotifyGroupID", slotifyGroupID))
			sendError(w, http.StatusInternalServerError, "get all slotifyGroup members query timed out")
			return
		default:
			s.Logger.Errorf("slotifyGroup api: failed to get all slotifyGroup members: %w", err)
			sendError(w, http.StatusInternalServerError,
				fmt.Sprintf(
					"slotifyGroup api: failed to get slotifyGroup members for slotifyGroup with id %d",
					slotifyGroupID))
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}

// (POST /api/slotify-groups/{slotifyGroupID}/users/me).
func (s Server) PostAPISlotifyGroupsSlotifyGroupIDUsersMe(w http.ResponseWriter,
	r *http.Request, slotifyGroupID uint32,
) {
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	s.PostAPISlotifyGroupsSlotifyGroupIDUsersUserID(w, r, slotifyGroupID, userID)
}

// (POST /api/slotify-groups/{slotifyGroupID}/users/{userID})
// nolint: funlen // TODO: Refactor this
func (s Server) PostAPISlotifyGroupsSlotifyGroupIDUsersUserID(w http.ResponseWriter,
	r *http.Request, slotifyGroupID uint32, userID uint32,
) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*database.DatabaseTimeout)
	defer cancel()

	rowsAffected, err := s.DB.AddUserToSlotifyGroup(ctx, database.AddUserToSlotifyGroupParams{
		UserID:         userID,
		SlotifyGroupID: slotifyGroupID,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: context cancelled",
				zap.Uint32("userID", userID), zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "slotifyGroup api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: add user to slotifyGroup query timed out",
				zap.Uint32("userID", userID), zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "slotifyGroup api: add slotifyGroup query timed out")
			return
		case database.IsRowDoesNotExistSQLError(err):
			s.Logger.Errorf("slotifyGroup api: user id or slotifyGroup id invalid fk: %w", err)
			sendError(w, http.StatusForbidden,
				fmt.Sprintf("slotifyGroup api: slotifyGroup id(%d) or user id(%d) was invalid", slotifyGroupID, userID))
			return
		default:
			s.Logger.Errorf("slotifyGroup api: failed to add user to slotifyGroup: %w", err)
			sendError(w, http.StatusInternalServerError, "slotifyGroup api: failed to add user to slotifyGroup")
			return
		}
	}

	if rowsAffected != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{1},
		}
		s.Logger.Error(zap.Uint32("userID", userID),
			zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "user api: failed to add member to slotifyGroup")
		return
	}

	t, err := s.DB.GetSlotifyGroupByID(ctx, slotifyGroupID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("slotifyGroup api: add user to slotifyGroup: get slotifyGroup by id: context cancelled",
				zap.Uint32("userID", userID), zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "slotifyGroup api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("slotifyGroup api: add user to slotifyGroup: get slotifyGroup by id query timed out",
				zap.Uint32("userID", userID), zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "slotifyGroup api: add slotifyGroup query timed out")
			return
		default:
			s.Logger.Errorf("slotifyGroup api: add user to slotifyGroup: get slotifyGroup by id query timed out",
				zap.Error(err))
			sendError(w, http.StatusInternalServerError,
				"slotifyGroup api: added slotifyGroup but failed to get slotifyGroup by id")
			return
		}
	}

	dbParams := database.GetAllSlotifyGroupMembersExceptParams{
		SlotifyGroupID: slotifyGroupID,
		UserID:         userID,
	}
	members, err := s.DB.GetAllSlotifyGroupMembersExcept(ctx, dbParams)
	if err != nil {
		s.Logger.Errorf(
			"slotifyGroup api: failed to get slotifyGroup members except new for PostAPISlotifyGroupsSlotifyGroupIDUsersUserID",
			zap.Uint32("exceptUserID", userID),
			zap.Error(err))
		// Still return ok because the actual endpoint worked, its just notifications that didnt
		SetHeaderAndWriteResponse(w, http.StatusOK, t)
		return
	}

	u, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		s.Logger.Errorf("slotifyGroup api: failed to get user details for PostAPISlotifyGroupsSlotifyGroupIDUsersUserID",
			zap.Error(err))
		// Still return ok because the actual endpoint worked, its just notifications that didnt
		SetHeaderAndWriteResponse(w, http.StatusCreated, t)
		return
	}
	allMemberNotif := database.CreateNotificationParams{
		Message: fmt.Sprintf("Say hi to %s, he just joined SlotifyGroup %s", u.FirstName+" "+u.LastName, t.Name),
		Created: time.Now(),
	}

	newMemberNotif := database.CreateNotificationParams{
		Message: fmt.Sprintf("You were added to SlotifyGroup %s!", t.Name),
		Created: time.Now(),
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, members, allMemberNotif)
	if err != nil {
		// Don't return error, attempt to send individual notification too
		s.Logger.Errorf(
			"slotifyGroup api: failed to send notification to all existing users of slotifyGroup, adding slotifyGroup member",
			zap.Error(err))
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, newMemberNotif)
	if err != nil {
		s.Logger.Errorf(
			"slotifyGroup api: failed to send notification to user that just joined slotifyGroup",
			zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, t)
}

// (OPTIONS /api/slotify-groups).
func (s Server) OptionsAPISlotifyGroups(w http.ResponseWriter, _ *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")   // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}
