package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/google/uuid"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request, params GetAPICalendarMeParams) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	//Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	//Create msgraph client
	graph, err := CreateMSFTGraphClient(ctx, s.Logger, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	//Make call to API route and parse events
	calendarEvents, err := makeCalendarMeAPICall(graph, params.Start, params.End)
	if err != nil {
		s.Logger.Error("failed to make calendar me msgraph api call", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to make calendar me msgraph api call")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}

// (POST /calendar/event).
// nolint: funlen // TODO: reduce length
func (s Server) PostAPICalendarMe(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.GetUserIDFromReq(r)
	if err != nil {
		s.Logger.Error("failed to get userid from request access token", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var eventRequest CalendarEvent
	if err = json.NewDecoder(r.Body).Decode(&eventRequest); err != nil {
		s.Logger.Error("failed to parse event body", zap.Error(err))
		sendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	at, err := getMSFTAccessToken(context.Background(), s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to get microsoft access token", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "Failed to get access token.")
		return
	}

	graphClient, err := createMSFTGraphClient(at)
	if err != nil {
		s.Logger.Error("failed to create Microsoft Graph client", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to create Microsoft Graph client")
		return
	}

	event := graphmodels.NewEvent()
	event.SetSubject(eventRequest.Subject)

	body := graphmodels.NewItemBody()
	contentType := graphmodels.HTML_BODYTYPE
	body.SetContentType(&contentType)
	body.SetContent(eventRequest.Body)
	event.SetBody(body)

	timeZone := "UTC"

	start := graphmodels.NewDateTimeTimeZone()
	start.SetDateTime(eventRequest.StartTime)
	start.SetTimeZone(&timeZone)
	event.SetStart(start)

	end := graphmodels.NewDateTimeTimeZone()
	end.SetDateTime(eventRequest.EndTime)
	end.SetTimeZone(&timeZone)
	event.SetEnd(end)

	// is location required and roomtype is not a property of location in graph
	location := graphmodels.NewLocation()
	location.SetDisplayName((*eventRequest.Locations)[0].Name)

	var attendees []graphmodels.Attendeeable
	for _, att := range *eventRequest.Attendees {
		attendee := graphmodels.NewAttendee()
		email := graphmodels.NewEmailAddress()
		mail := string(*att.Email)
		email.SetAddress(&mail) // this feel jank
		attendee.SetEmailAddress(email)

		var attendeeType graphmodels.AttendeeType
		if att.Type != nil {
			switch *att.Type {
			case Required:
				attendeeType = graphmodels.REQUIRED_ATTENDEETYPE
			case Optional:
				attendeeType = graphmodels.OPTIONAL_ATTENDEETYPE
			case Resource:
				attendeeType = graphmodels.RESOURCE_ATTENDEETYPE
			default:
				attendeeType = graphmodels.REQUIRED_ATTENDEETYPE
			}
		}
		attendee.SetTypeEscaped(&attendeeType)

		// response status?
		responseStatus := graphmodels.NewResponseStatus()
		response := graphmodels.NOTRESPONDED_RESPONSETYPE
		responseStatus.SetResponse(&response)
		attendees = append(attendees, attendee)
	}

	event.SetAttendees(attendees)

	transactionID := uuid.New().String()
	event.SetTransactionId(&transactionID)

	events, err := graphClient.Me().Events().Post(context.Background(), event, nil)
	if err != nil {
		s.Logger.Error("failed to create calendar event", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to create event")
		return
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
