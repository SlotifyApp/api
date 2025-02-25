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
	"go.uber.org/zap"
)

// (POST /api/invites) Create a new invite.
func (s Server) PostAPIInvites(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var invitesCreateBody PostAPIInvitesJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&invitesCreateBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Object("body", invitesCreateBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	params := database.CreateInviteParams{
		SlotifyGroupID: invitesCreateBody.SlotifyGroupID,
		FromUserID:     userID,
		ToUserID:       invitesCreateBody.ToUserID,
		Message:        invitesCreateBody.Message,
		CreatedAt:      invitesCreateBody.CreatedAt,
	}

	err = retry.Do(func() error {
		var rows int64
		rows, err = s.DB.CreateInvite(ctx, params)
		if rows != 1 {
			return database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
		}
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("context cancelled creating invite: %w",
					err)
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("deadline exceeded during creating invite: %w", err)
			default:
				return fmt.Errorf("failed to create invite: %w", err)
			}
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("failed to create invite", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create invite")
		return
	}

	var g database.SlotifyGroup
	if g, err = s.DB.GetSlotifyGroupByID(ctx, invitesCreateBody.SlotifyGroupID); err != nil {
		s.Logger.Errorf("invite api: failed to send notification to toUser",
			zap.Error(err))
		SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully created invite")
		return
	}

	// Create notification to user who has been invited
	if err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB,
		[]uint32{invitesCreateBody.ToUserID}, database.CreateNotificationParams{
			Message: fmt.Sprintf("You have a new invite to team %s!", g.Name),
			Created: time.Now(),
		}); err != nil {
		s.Logger.Errorf("invite api: failed to send notification to toUser",
			zap.Error(err))
	}

	var u database.User
	if u, err = s.DB.GetUserByID(ctx, invitesCreateBody.ToUserID); err != nil {
		s.Logger.Errorf("invite api: failed to send notification to fromUser",
			zap.Error(err))
		SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully created invite")
		return
	}

	// Create notification to user who created the invite
	if err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB,
		[]uint32{userID}, database.CreateNotificationParams{
			Message: fmt.Sprintf(
				"You successfully created an invite on behalf of team %s to %s %s!",
				g.Name, u.FirstName, u.LastName), Created: time.Now(),
		}); err != nil {
		s.Logger.Errorf("invite api: failed to send notification to fromUser",
			zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully created invite")
}

// (GET /api/invites/me Get all invites for logged in user.)
func (s Server) GetAPIInvitesMe(w http.ResponseWriter, r *http.Request, params GetAPIInvitesMeParams) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	invites, err := s.DB.ListInvitesMe(ctx, database.ListInvitesMeParams{Status: params.Status, ToUserID: userID})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("invite api: failed to get invites: context cancelled")
			sendError(w, http.StatusInternalServerError, "user api: failed to get invites")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("invite api: failed to get invites: query timed out")
			sendError(w, http.StatusInternalServerError, "invite api: failed to get invites")
			return
		default:
			s.Logger.Error("invite api: failed to get invites")
			sendError(w, http.StatusInternalServerError, "user api: failed to get invites")
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, invites)
}

// (DELETE /api/invites/{inviteID}Delete an invite).
func (s Server) DeleteAPIInvitesInviteID(w http.ResponseWriter, r *http.Request, inviteID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	var err error
	err = retry.Do(func() error {
		var rows int64
		rows, err = s.DB.DeleteInviteByID(ctx, inviteID)

		if rows != 1 {
			return database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("context cancelled deleting invite: %w",
					err)
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("deadline exceeded during deleting invite: %w", err)
			default:
				return fmt.Errorf("failed to delete invite: %w", err)
			}
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("failed to delete invite", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create invite")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully deleted invite!")
}

// (PATCH /api/invites/{inviteID} update a new invite with a new message).
func (s Server) PatchAPIInvitesInviteID(w http.ResponseWriter, r *http.Request, inviteID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body PatchAPIInvitesInviteIDJSONRequestBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Object("body", body), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
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
				return fmt.Errorf("context cancelled deleting invite: %w",
					err)
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("deadline exceeded during deleting invite: %w", err)
			default:
				return fmt.Errorf("failed to delete invite: %w", err)
			}
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("failed to update invite message", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update invite message")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully updated invite message!")
}

// (PATCH /api/invites/{inviteID}/status/{newStatus} Update a new invite).
func (s Server) PatchAPIInvitesInviteIDStatusNewStatus(w http.ResponseWriter, r *http.Request,
	inviteID uint32, newStatus InviteStatus,
) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	var err error
	err = retry.Do(func() error {
		var rows int64
		rows, err = s.DB.UpdateInviteStatus(ctx,
			database.UpdateInviteStatusParams{
				ID:     inviteID,
				Status: database.InviteStatus(newStatus),
			})

		if rows != 1 {
			return database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("context cancelled updating invite status: %w",
					err)
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("deadline exceeded during updating invite status: %w", err)
			default:
				return fmt.Errorf("failed to update invite status: %w", err)
			}
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("failed to update invite", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update invite")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully updated invite status!")
}

// (GET /api/slotify-groups/{slotifyGroupID}/invites Get all invites for a slotify group).
func (s Server) GetAPISlotifyGroupsSlotifyGroupIDInvites(w http.ResponseWriter,
	r *http.Request, slotifyGroupID uint32, params GetAPISlotifyGroupsSlotifyGroupIDInvitesParams,
) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*database.DatabaseTimeout)
	defer cancel()

	invites, err := s.DB.ListInvitesByGroup(ctx,
		database.ListInvitesByGroupParams{
			SlotifyGroupID: slotifyGroupID,
			Status:         params.Status,
		})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("context cancelled getting invites group", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to update invite")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("context deadline exceeded while getting invites group", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to update invite")
			return
		default:
			s.Logger.Error("failed to get group invites", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get group invites")
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, invites)
}

// (OPTIONS /api/invites).
func (s Server) OptionsAPIInvites(w http.ResponseWriter, _ *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")   // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}

// (OPTIONS /api/invites/{inviteID}Satisfy CORS preflight for invites.)
func (s Server) OptionsAPIInvitesInviteID(w http.ResponseWriter, _ *http.Request, _ uint32) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")   // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}

// (OPTIONS /api/invites/{inviteID}/status/{newStatus}Satisfy CORS preflight for creating invites.)
func (s Server) OptionsAPIInvitesInviteIDStatusNewStatus(w http.ResponseWriter,
	_ *http.Request, _ uint32, _ InviteStatus,
) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")   // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}
