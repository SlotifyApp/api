package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

// (DELETE /api/slotify-groups/{slotifyGroupID}/leave/me)  Have a member leave from a slotify group.
func (s Server) DeleteSlotifyGroupsSlotifyGroupIDLeaveMe(w http.ResponseWriter,
	r *http.Request, slotifyGroupID uint32,
) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var count int64
	var err error
	if count, err = s.DB.CheckMemberInSlotifyGroup(ctx, database.CheckMemberInSlotifyGroupParams{
		UserID:         userID,
		SlotifyGroupID: slotifyGroupID,
	}); err != nil {
		s.Logger.Error("failed to get count of member in slotify group", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to check if member is in slotify group")
		return
	}

	// User not a member of the group
	if count == 0 {
		s.Logger.Error("user was not a member of the group, so can't leave", zap.Error(err),
			zap.Uint32("userID", userID),
			zap.Uint32("slotifyGroupID", slotifyGroupID),
		)
		sendError(w, http.StatusUnauthorized, "You are not a member of the group, so you can't leave")
		return
	}

	var memberCount int64
	if memberCount, err = s.DB.CountSlotifyGroupMembers(ctx, slotifyGroupID); err != nil {
		s.Logger.Error("failed to get member count", zap.Error(err),
			zap.Uint32("userID", userID),
			zap.Uint32("slotifyGroupID", slotifyGroupID),
		)
		sendError(w, http.StatusInternalServerError, "Failed to leave slotify group")
		return
	}

	// This user is the only member of the group, so also delete the group
	if memberCount == 1 {
		var rowsAffected int64
		if rowsAffected, err = s.DB.DeleteSlotifyGroupByID(ctx, slotifyGroupID); err != nil {
			s.Logger.Error("failed to delete slotify group", zap.Error(err),
				zap.Uint32("userID", userID),
				zap.Uint32("slotifyGroupID", slotifyGroupID),
			)
			sendError(w, http.StatusInternalServerError, "Failed to leave slotify group")
			return
		}

		var expectedRows int64 = 2
		if rowsAffected != expectedRows {
			err = database.WrongNumberSQLRowsError{ActualRows: rowsAffected, ExpectedRows: []int64{expectedRows}}
			s.Logger.Error("failed to delete slotify group", zap.Error(err),
				zap.Uint32("userID", userID),
				zap.Uint32("slotifyGroupID", slotifyGroupID),
			)
			sendError(w, http.StatusInternalServerError, "Failed to leave slotify group")
			return
		}
	} else {
		var rowsAffected int64
		if rowsAffected, err = s.DB.RemoveSlotifyGroupMember(ctx, database.RemoveSlotifyGroupMemberParams{
			UserID:         userID,
			SlotifyGroupID: slotifyGroupID,
		}); err != nil {
			err = database.WrongNumberSQLRowsError{ActualRows: rowsAffected, ExpectedRows: []int64{1}}
			s.Logger.Error("failed to remove slotify member", zap.Error(err),
				zap.Uint32("userID", userID),
				zap.Uint32("slotifyGroupID", slotifyGroupID),
			)
			sendError(w, http.StatusInternalServerError, "Failed to leave slotify group")
			return
		}
	}

	p := sendLeaverNotificationsParams{
		ctx:            ctx,
		slotifyGroupID: slotifyGroupID,
		userID:         userID,
		l:              s.Logger,
		db:             s.DB,
		notifService:   s.NotificationService,
	}
	// Still successful, just notifications failed
	if err = sendLeaverNotifications(p); err != nil {
		s.Logger.Error("failed to send leaver notifications", zap.Error(err),
			zap.Uint32("userID", userID),
			zap.Uint32("slotifyGroupID", slotifyGroupID),
		)
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully left the group.")
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

	tx, err := s.DB.DB.Begin()
	if err != nil {
		s.Logger.Error("failed to start db transaction", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "callback route: failed to start db transaction")
		return
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			s.Logger.Error("failed to rollback db transaction", zap.Error(err))
		}
	}()

	qtx := s.DB.WithTx(tx)

	var slotifyGroupID int64
	slotifyGroupID, err = qtx.AddSlotifyGroup(ctx, slotifyGroupBody.Name)
	if err != nil {
		switch {
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

	p := AddUserToSlotifyGroupParams{
		ctx:    ctx,
		userID: userID,
		//nolint: gosec // id is unsigned 32 bit int
		slotifyGroupID: uint32(slotifyGroupID),
		l:              s.Logger,
		qtx:            qtx,
		notifService:   s.NotificationService,
	}

	if err = AddUserToSlotifyGroup(p); err != nil {
		s.Logger.Error("creating group was successful, but adding user to the group was not",
			zap.Uint32("userID", userID),
			zap.Int64("slotifyGroupID", slotifyGroupID),
			zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to create a group and add you to it")
		return
	}

	if err = tx.Commit(); err != nil {
		s.Logger.Error("failed to commit db transaction", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to create a group and add you to it")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, SlotifyGroup{
		//nolint: gosec // id is unsigned 32 bit int
		Id:   uint32(slotifyGroupID),
		Name: slotifyGroupBody.Name,
	})
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
		s.Logger.Error("slotifyGroup api: failed to get users of a group: failed to get count of group by id",
			zap.Uint32("slotifyGroupID",
				slotifyGroupID),
			zap.Error(err),
		)
		sendError(w, http.StatusInternalServerError, "sorry try again")
		return
	}

	if count == 0 {
		s.Logger.Error("slotifyGroup api: slotifyGroup members of non-existent slotifyGroup requested",
			zap.Uint32("slotifyGroupID", slotifyGroupID), zap.Error(err))
		sendError(w, http.StatusNotFound,
			fmt.Sprintf("slotifyGroup api: slotifyGroup with id(%d) does not exist", slotifyGroupID))
		return
	}

	users, err := s.DB.GetAllSlotifyGroupMembers(ctx, slotifyGroupID)
	if err != nil {
		s.Logger.Error("slotifyGroup api: failed to get users of a group",
			zap.Uint32("slotifyGroupID",
				slotifyGroupID),
			zap.Error(err),
		)
		sendError(w, http.StatusInternalServerError, "failed to get all slotify group members")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}
