package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/google/uuid"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.GetUserIDFromReq(r)
	if err != nil {
		s.Logger.Error("failed to get userid from request access token")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	at, err := getMSFTAccessToken(context.Background(), s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to get microsoft access token", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := createMSFTGraphClient(at)

	if err != nil || graph == nil {
		s.Logger.Error("failed to create msft graph client", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "msft graph could not be created")
		return
	}

	events, err := graph.Me().Calendar().Events().Get(context.Background(), nil)
	if err != nil {
		s.Logger.Error("failed to call graph client route", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to call graph client route")
		return
	}
	calendarEvents := []CalendarEvent{}
	for _, e := range events.GetValue() {
		ce := CalendarEvent{
			StartTime: e.GetStart().GetDateTime(),
			EndTime:   e.GetEnd().GetDateTime(),
			Subject:   e.GetSubject(),
		}
		calendarEvents = append(calendarEvents, ce)
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}

// (POST /calendar/event).
func (s Server) PostAPICalendarCreate(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.GetUserIDFromReq(r)
	if err != nil {
		s.Logger.Error("failed to get userid from request access token", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var eventRequest CalendarEvent
	if err := json.NewDecoder(r.Body).Decode(&eventRequest); err != nil {
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
			case "required":
				attendeeType = graphmodels.REQUIRED_ATTENDEETYPE
			case "optional":
				attendeeType = graphmodels.OPTIONAL_ATTENDEETYPE
			case "resource":
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

	transactionId := uuid.New().String()
	event.SetTransactionId(&transactionId)

	events, err := graphClient.Me().Events().Post(context.Background(), event, nil)
	if err != nil {
		s.Logger.Error("failed to create calendar event", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, events)
}
