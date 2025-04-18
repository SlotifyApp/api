package api

import (
	"context"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

// (GET /api/events), HTTP SSE route.
func (s Server) RenderEvent(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Check if the response writer implements the flush interface
	// if so, flush our event-stream headers
	f, ok := w.(http.Flusher)

	if !ok {
		logger.Error("failed to flush event stream headers, responsewriter did not implement Flusher interface")
		sendError(w, http.StatusInternalServerError, "failed to flush event stream headers")
		return
	}

	f.Flush()

	if err := s.NotificationService.RegisterUserClient(s.Logger, userID, w); err != nil {
		logger.Errorf("failed to register user client", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "failed to register user client")
		return
	}

	// Block until the request is 'done', eg. client navigates away
	<-r.Context().Done()

	logger.Infof("userID %d disconnected", userID)
	s.NotificationService.DeleteUserConn(s.Logger, userID, w)
}

// (PATCH /api/notifications/{notificationID}/read). Mark notifications as being read.
func (s Server) PatchAPINotificationsNotificationIDRead(w http.ResponseWriter, r *http.Request, notificationID uint32) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	dbParams := database.MarkNotificationAsReadParams{
		UserID:         userID,
		NotificationID: notificationID,
	}

	rowsAffected, err := s.DB.MarkNotificationAsRead(ctx, dbParams)
	if err != nil {
		logger.Error("failed to mark notification as read", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to mark notification as read.")
		return
	}

	if rowsAffected != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{1},
		}
		logger.Error("failed to mark notification as read", zap.Error(err))
		sendError(w, http.StatusBadRequest, "Failed to mark notification as read.")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Marked notification as read.")
}

// (GET /api/notifications/{notificationID}/read).
func (s Server) GetAPIUsersMeNotifications(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	notifs, err := s.DB.GetUnreadUserNotifications(ctx, userID)
	if err != nil {
		logger.Error("failed to get unread user notifications")
		sendError(w, http.StatusInternalServerError, "failed to get unread user notifications from db")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, notifs)
}
