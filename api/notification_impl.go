package api

import (
	"context"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

// (GET /api/events), HTTP SSE route.
func (s Server) RenderEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Allow frontend origin
	w.Header().Set("Access-Control-Allow-Credentials", "true")             // Allow cookies to be sent
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Check if the response writer implements the flush interface
	// if so, flush our event-stream headers
	f, ok := w.(http.Flusher)

	if !ok {
		s.Logger.Error("failed to flush event stream headers, responsewriter did not implement Flusher interface")
		sendError(w, http.StatusInternalServerError, "failed to flush event stream headers")
		return
	}

	f.Flush()

	if err := s.NotificationService.RegisterUserClient(s.Logger, userID, w); err != nil {
		s.Logger.Errorf("failed to register user client", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "failed to register user client")
		return
	}

	// Block until the request is 'done', eg. client navigates away
	<-r.Context().Done()

	s.Logger.Infof("userID %d disconnected", userID)
	s.NotificationService.DeleteUserConn(s.Logger, userID, w)
}

// (OPTIONS /api/notifications/{notificationID}/read).
func (s Server) OptionsAPINotificationsNotificationIDRead(w http.ResponseWriter, _ *http.Request, _ uint32) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")   // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}

// (PATCH /api/notifications/{notificationID}/read). Mark notifications as being read.
func (s Server) PatchAPINotificationsNotificationIDRead(w http.ResponseWriter, r *http.Request, notificationID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	dbParams := database.MarkNotificationAsReadParams{
		UserID:         userID,
		NotificationID: notificationID,
	}

	rowsAffected, err := s.DB.MarkNotificationAsRead(ctx, dbParams)
	if err != nil {
		s.Logger.Error("failed to mark notification as read", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to mark notification as read.")
		return
	}

	if rowsAffected != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{1},
		}
		s.Logger.Error("failed to mark notification as read", zap.Error(err))
		sendError(w, http.StatusBadRequest, "Failed to mark notification as read.")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Marked notification as read.")
}

// (GET /api/notifications/{notificationID}/read).
func (s Server) GetAPIUsersMeNotifications(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	notifs, err := s.DB.GetUnreadUserNotifications(ctx, userID)
	if err != nil {
		s.Logger.Error("failed to get unread user notifications")
		sendError(w, http.StatusInternalServerError, "failed to get unread user notifications from db")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, notifs)
}
