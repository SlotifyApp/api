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
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// (GET /api/calendar/{userID}).
func (s Server) GetAPICalendarUserID(w http.ResponseWriter,
	r *http.Request, userID uint32, params GetAPICalendarUserIDParams,
) {
	loggedInUserID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", loggedInUserID))

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	// create graph client for the userID in query params.
	// TODO: stop private events etc. from being shown
	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Make call to API route and parse events
	calendarEvents, err := makeCalendarMeAPICall(graph, params.Start, params.End)
	if err != nil {
		logger.Error("failed to make calendar me msgraph api call", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to make calendar me msgraph api call")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request, params GetAPICalendarMeParams) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)

	s.GetAPICalendarUserID(w, r, userID, GetAPICalendarUserIDParams(params))
}

// (POST /calendar/me).
func (s Server) PostAPICalendarMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	var eventRequest CalendarEvent
	if err := json.NewDecoder(r.Body).Decode(&eventRequest); err != nil {
		logger.Error("failed to parse event body", zap.Error(ErrUnmarshalBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	event := parseCalendarEventToMSFTEvent(eventRequest)

	var createdEventable graphmodels.Eventable
	err = retry.Do(func() error {
		createdEventable, err = graph.Me().Events().Post(ctx, event, nil)
		if err != nil {
			return fmt.Errorf("graph api create calendar failed: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		logger.Error("failed to create calendar event", zap.Error(err))
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
		logger.Error("failed to send notification", zap.Error(err))
	}

	var parsedEvent []CalendarEvent
	if parsedEvent, err = parseEventableResp([]graphmodels.Eventable{createdEventable}); err != nil {
		logger.Error("failed to parse msft event response", zap.Error(err))
		sendError(w, http.StatusBadRequest, "failed to parse msft event response")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, parsedEvent[0])
}

// (GET /api/calendar/event).
func (s Server) GetAPICalendarEvent(w http.ResponseWriter, r *http.Request, params GetAPICalendarEventParams) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	// Make call to API route and parse events
	// Get old meeting data from microsoft
	var msftMeeting graphmodels.Eventable
	//nolint: nestif // nesting complexity is not too much
	if params.IsICalUId {
		queryFilter := "iCalUId eq '" + params.MsftID + "'"

		// Get old meeting data from microsoft
		requestConfig := users.ItemEventsRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemEventsRequestBuilderGetQueryParameters{
				Filter: &queryFilter,
			},
		}

		meetingObj, err := s.DB.GetMeetingByMSFTID(ctx, params.MsftID)
		if err != nil {
			logger.Error("meeting not found in db to find owner of meeting with iCalUID", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Meeting not found in db to find owner of meeting")
			return
		}

		// Get owner's user id
		ownerObj, err := s.DB.GetUserByEmail(ctx, meetingObj.OwnerEmail)
		if err != nil {
			logger.Error("failed to get owner obj from db using email: ", zap.Error(err))
			sendError(w, http.StatusBadGateway, "owner is not found as a slotify user")
			return
		}

		graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, ownerObj.ID)
		if err != nil {
			logger.Error("failed to create msgraph client with owner id", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
			return
		}

		var msftMeetingRes graphmodels.EventCollectionResponseable
		msftMeetingRes, err = graph.Me().Events().Get(ctx, &requestConfig)
		if err != nil {
			logger.Error("failed to get meeting data from microsoft", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get meeting data from microsoft")
			return
		}

		if msftMeetingRes != nil && msftMeetingRes.GetValue() != nil &&
			len(msftMeetingRes.GetValue()) > 0 {
			msftMeeting = msftMeetingRes.GetValue()[0]
		} else {
			logger.Error("failed to get meeting data from microsoft", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get meeting data from microsoft")
			return
		}
	} else {
		// create graph client for the userID in query params.
		graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
		if err != nil {
			logger.Error("failed to create msgraph client", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
			return
		}

		msftMeeting, err = graph.Me().Events().ByEventId(params.MsftID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to get meeting data from microsoft", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get meeting data from microsoft")
			return
		}
	}

	var parsedEvents []CalendarEvent
	var err error
	if parsedEvents, err = parseEventableResp([]graphmodels.Eventable{msftMeeting}); err != nil {
		logger.Error("failed to get meeting data from microsoft", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get meeting data from microsoft")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, parsedEvents[0])
}
