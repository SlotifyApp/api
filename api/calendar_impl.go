package api

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request, params GetAPICalendarMeParams) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	userID, ok := r.Context().Value(UserCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	at, err := getMSFTAccessToken(ctx, s.MSALClient, s.DB, userID)
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

	start := params.Start.Format(time.RFC3339)
	end := params.End.Format(time.RFC3339)

	requestParameters := &graphusers.ItemCalendarCalendarViewRequestBuilderGetQueryParameters{
		StartDateTime: &start,
		EndDateTime:   &end,
	}

	configuration := &graphusers.ItemCalendarCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	events, err := graph.Me().Calendar().CalendarView().Get(context.Background(), configuration)
	if err != nil {
		s.Logger.Error("failed to call graph client route", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to call graph client route")
		return
	}

	calendarEvents := []CalendarEvent{}
	for _, e := range events.GetValue() {
		attendees := parseMSFTAttendees(e)
		locations := parseMSFTLocations(e)

		var joinURL *string
		if e.GetOnlineMeeting() != nil {
			joinURL = e.GetOnlineMeeting().GetJoinUrl()
		}

		var endTime *string
		if e.GetEnd() != nil {
			endTime = e.GetEnd().GetDateTime()
		}

		var startTime *string
		if e.GetStart() != nil {
			startTime = e.GetStart().GetDateTime()
		}

		ce := CalendarEvent{
			Attendees:   &attendees,
			Body:        e.GetBodyPreview(),
			Created:     e.GetCreatedDateTime(),
			EndTime:     endTime,
			Id:          e.GetId(),
			IsCancelled: e.GetIsCancelled(),
			JoinURL:     joinURL,
			Locations:   &locations,
			Organizer:   (*openapi_types.Email)(e.GetOrganizer().GetEmailAddress().GetAddress()),
			StartTime:   startTime,
			Subject:     e.GetSubject(),
			WebLink:     e.GetWebLink(),
		}
		calendarEvents = append(calendarEvents, ce)
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}
