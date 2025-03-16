package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/avast/retry-go"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

const (
	InvitesLimitMax = 50
)

// (POST /api/invites) Create a new invite.
func (s Server) PostAPIInvites(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 4*database.DatabaseTimeout)
	defer cancel()

	var err error
	var invitesCreateBody PostAPIInvitesJSONRequestBody
	if err = json.NewDecoder(r.Body).Decode(&invitesCreateBody); err != nil {
		logger.Error(ErrUnmarshalBody, zap.Object("body", invitesCreateBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	var g database.SlotifyGroup
	if g, err = s.DB.GetSlotifyGroupByID(ctx, invitesCreateBody.SlotifyGroupID); err != nil {
		logger.Errorf("invite api: failed to get group by id", zap.Error(err))
		sendError(w, http.StatusBadRequest, "failed to get group by id")
		return
	}

	var u database.User
	if u, err = s.DB.GetUserByID(ctx, userID); err != nil {
		logger.Errorf("invite api: failed to get user by id", zap.Error(err))
		sendError(w, http.StatusBadRequest, "failed to get user by id")
		return
	}

	var toUser database.User
	if toUser, err = s.DB.GetUserByID(ctx, invitesCreateBody.ToUserID); err != nil {
		logger.Errorf("invite api: failed to get user by id", zap.Error(err))
		sendError(w, http.StatusBadRequest, "failed to get user by id")
		return
	}

	// check if fromUser is in group and check if toUser is in group
	if err = checkIfUsersInGroup(checkIfUsersInGroupParams{
		ctx:              ctx,
		db:               s.DB,
		fromUserID:       userID,
		toUserFirstName:  toUser.FirstName,
		toUserLastName:   toUser.LastName,
		slotifyGroupID:   invitesCreateBody.SlotifyGroupID,
		slotifyGroupName: g.Name,
		toUserID:         invitesCreateBody.ToUserID,
	}); err != nil {
		logger.Errorf("invite api: ", zap.Error(err))
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	params := database.CreateInviteParams{
		SlotifyGroupID: invitesCreateBody.SlotifyGroupID,
		FromUserID:     userID,
		ToUserID:       invitesCreateBody.ToUserID,
		Message:        invitesCreateBody.Message,
		ExpiryDate:     invitesCreateBody.ExpiryDate.Time,
		Status:         database.InviteStatusPending,
		CreatedAt:      invitesCreateBody.CreatedAt,
	}

	var inviteID int64
	err = retry.Do(func() error {
		if inviteID, err = s.DB.CreateInvite(ctx, params); err != nil {
			return fmt.Errorf("failed to create invite: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		logger.Error("failed to create invite", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create invite")
		return
	}

	sendPostInviteNotification(sendPostInviteNotificationParams{
		ctx:             ctx,
		toUserID:        invitesCreateBody.ToUserID,
		fromUserID:      userID,
		notifService:    s.NotificationService,
		logger:          s.Logger,
		db:              s.DB,
		groupName:       g.Name,
		toUserFirstName: toUser.FirstName,
		toUserLastName:  toUser.LastName,
	})

	createdInvite := InvitesGroup{
		CreatedAt:         invitesCreateBody.CreatedAt,
		ExpiryDate:        openapi_types.Date{Time: invitesCreateBody.ExpiryDate.Time},
		FromUserEmail:     openapi_types.Email(u.Email),
		FromUserFirstName: u.FirstName, FromUserLastName: u.LastName,
		//nolint: gosec // id is unsigned 32 bit int
		InviteID: uint32(inviteID), Message: invitesCreateBody.Message,
		Status: InviteStatusPending, ToUserEmail: openapi_types.Email(toUser.Email),
		ToUserFirstName: toUser.FirstName, ToUserLastName: toUser.LastName,
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, createdInvite)
}

// (GET /api/invites/me Get all invites for logged in user.)
func (s Server) GetAPIInvitesMe(w http.ResponseWriter, r *http.Request, params GetAPIInvitesMeParams) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 4*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	invitesLimit := min(params.Limit, InvitesLimitMax)

	//nolint: gosec // page is unsigned 32 bit int
	lastID := uint32(*params.PageToken)

	invites, err := s.DB.ListInvitesMe(ctx, database.ListInvitesMeParams{
		Status:   params.Status,
		ToUserID: userID,
		LastID:   lastID,
		Limit:    invitesLimit,
	})
	if err != nil {
		logger.Error("invite api: failed to get invites", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "user api: failed to get invites")
		return
	}

	var nextPageToken int
	if len(invites) == int(invitesLimit) {
		nextPageToken = int(invites[len(invites)-1].InviteID)
	} else {
		nextPageToken = -1
	}

	response := struct {
		Invites       []database.ListInvitesMeRow
		NextPageToken int
	}{
		Invites:       invites,
		NextPageToken: nextPageToken,
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, response)
}

// (DELETE /api/invites/{inviteID} Delete an invite).
func (s Server) DeleteAPIInvitesInviteID(w http.ResponseWriter, r *http.Request, inviteID uint32) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	var invite database.Invite
	var err error
	if invite, err = s.DB.GetInviteByID(ctx, inviteID); err != nil {
		logger.Error("failed to get invite by id", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get invite by id")
		return
	}

	var userIsInGroup bool
	if userIsInGroup, err = database.CheckMemberInSlotifyGroupWrapper(ctx, s.DB, database.CheckMemberInSlotifyGroupParams{
		UserID:         userID,
		SlotifyGroupID: invite.SlotifyGroupID,
	}); err != nil {
		logger.Error("failed to see if user is in group", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to see if user is in group")
		return
	}

	if !userIsInGroup {
		logger.Error("user is not in group, cannot delete invite", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "You are not apart of the group, cannto delete invite.")
		return
	}

	err = retry.Do(func() error {
		return database.DeleteInviteByIDWrapper(ctx, s.DB, inviteID)
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		logger.Error("failed to delete invite", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create invite")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully deleted invite!")
}

// (PATCH /api/invites/{inviteID} update a new invite with a new message).
func (s Server) PatchAPIInvitesInviteID(w http.ResponseWriter, r *http.Request, inviteID uint32) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	var body PatchAPIInvitesInviteIDJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error(ErrUnmarshalBody, zap.Object("body", body), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	var invite database.Invite
	if invite, err = s.DB.GetInviteByID(ctx, inviteID); err != nil {
		logger.Error("failed to get invite details from invite id", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to get invite details from invite id")
		return
	}

	if invite.FromUserID != userID {
		logger.Error("user cannot", zap.Error(err))
		sendError(w, http.StatusUnauthorized,
			"can only edit your invite message, contact the person who created the invite")
		return
	}

	err = retry.Do(func() error {
		var rows int64
		rows, err = s.DB.UpdateInviteMessage(ctx,
			database.UpdateInviteMessageParams{
				FromUserID: userID,
				ID:         inviteID,
				Message:    body.Message,
			})

		if rows != 1 {
			return database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("context cancelled deleting invite: %w", err)
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("deadline exceeded during deleting invite: %w", err)
			default:
				return fmt.Errorf("failed to delete invite: %w", err)
			}
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		logger.Error("failed to update invite message", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update invite message")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully updated invite message!")
}

// (PATCH /api/invites/{inviteID}/decline Decline an invite).
func (s Server) PatchAPIInvitesInviteIDDecline(w http.ResponseWriter, r *http.Request, inviteID uint32) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	p := validateAndUpdateInviteStatusParams{
		ctx:       ctx,
		qtx:       &s.DB.Queries,
		inviteID:  inviteID,
		l:         s.Logger,
		userID:    userID,
		newStatus: InviteStatusAccepted,
	}

	_, err := validateAndUpdateInviteStatus(p)
	if err != nil {
		logger.Error("failed to validate and update invite status", zap.Error(err))
		sendError(w, http.StatusBadGateway, err.Error())
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully declined invite.")
}

// (PATCH /api/invites/{inviteID}/accept Accept an invite).
func (s Server) PatchAPIInvitesInviteIDAccept(w http.ResponseWriter, r *http.Request, inviteID uint32) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	tx, err := s.DB.DB.Begin()
	if err != nil {
		logger.Error("failed to start db transaction", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "callback route: failed to start db transaction")
		return
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			logger.Error("failed to rollback db transaction", zap.Error(err))
		}
	}()

	qtx := s.DB.WithTx(tx)

	p := validateAndUpdateInviteStatusParams{
		ctx:       ctx,
		qtx:       qtx,
		inviteID:  inviteID,
		l:         s.Logger,
		userID:    userID,
		newStatus: InviteStatusAccepted,
	}

	var invite database.Invite
	if invite, err = validateAndUpdateInviteStatus(p); err != nil {
		logger.Error("failed to validate and update invite status", zap.Error(err))
		sendError(w, http.StatusBadGateway, err.Error())
		return
	}

	addUserParams := AddUserToSlotifyGroupParams{
		ctx:            ctx,
		userID:         userID,
		slotifyGroupID: invite.SlotifyGroupID,
		l:              s.Logger,
		qtx:            qtx,
		notifService:   s.NotificationService,
	}

	if err = AddUserToSlotifyGroup(addUserParams); err != nil {
		logger.Error("failed to add user to slotify group", zap.Error(err),
			zap.Uint32("slotifyGroupID", invite.SlotifyGroupID),
			zap.Uint32("userID", userID),
		)
		sendError(w, http.StatusBadGateway, "failed to add you to the group, maybe you are already a member?")
		return
	}

	if err = tx.Commit(); err != nil {
		logger.Error("failed to commit db transaction", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to accept invite")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully accepted invite!")
}

// (GET /api/slotify-groups/{slotifyGroupID}/invites Get all invites for a slotify group).
func (s Server) GetAPISlotifyGroupsSlotifyGroupIDInvites(w http.ResponseWriter,
	r *http.Request, slotifyGroupID uint32, params GetAPISlotifyGroupsSlotifyGroupIDInvitesParams,
) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	invitesLimit := min(InvitesLimitMax, params.Limit)

	//nolint: gosec // page is unsigned 32 bit int
	lastID := uint32(*params.PageToken)

	invites, err := s.DB.ListInvitesByGroup(ctx,
		database.ListInvitesByGroupParams{
			SlotifyGroupID: slotifyGroupID,
			Status:         params.Status,
			LastID:         lastID,
			Limit:          invitesLimit,
		})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logger.Error("context cancelled getting invites group", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to update invite")
			return
		case errors.Is(err, context.DeadlineExceeded):
			logger.Error("context deadline exceeded while getting invites group", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to update invite")
			return
		default:
			logger.Error("failed to get group invites", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get group invites")
			return
		}
	}

	var nextPageToken int
	if len(invites) == int(invitesLimit) {
		nextPageToken = int(invites[len(invites)-1].InviteID)
	} else {
		nextPageToken = -1
	}

	response := struct {
		Invites       []database.ListInvitesByGroupRow `json:"invites"`
		NextPageToken int                              `json:"nextpageToken"`
	}{
		Invites:       invites,
		NextPageToken: nextPageToken,
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, response)
}
