package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/avast/retry-go"
	"go.uber.org/zap"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request, params GetAPICalendarMeParams) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	// Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Make call to API route and parse events
	calendarEvents, err := makeCalendarMeAPICall(graph, params.Start, params.End)
	if err != nil {
		s.Logger.Error("failed to make calendar me msgraph api call", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to make calendar me msgraph api call")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}

// (POST /calendar/event).
func (s Server) PostAPICalendarMe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var eventRequest CalendarEvent
	if err := json.NewDecoder(r.Body).Decode(&eventRequest); err != nil {
		s.Logger.Error("failed to parse event body", zap.Error(ErrUnmarshalBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	event := parseCalendarEventToMSFTEvent(eventRequest)

	var events graphmodels.Eventable
	err = retry.Do(func() error {
		events, err = graph.Me().Events().Post(ctx, event, nil)
		if err != nil {
			return fmt.Errorf("graph api create calendar failed: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("failed to create calendar event", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create event")
		return
	}

	// if success, send through notifications
	// TODO: send separate notification for other participants, saying a meeting has been added
	notif := database.CreateNotificationParams{
		Message: "Created meeting!",
		Created: time.Now(),
	}
	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, notif)
	if err != nil {
		// dont send http error because the actual op succeeded
		s.Logger.Error("failed to send notification", zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, events)
}

func (s Server) OptionsAPICalendarMe(w http.ResponseWriter, _ *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")        // Your frontend's origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // Allowed methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                    // Allow credentials (cookies, etc.)

	// Send a 204 No Content response to indicate that the preflight request was successful
	w.WriteHeader(http.StatusNoContent)
}
