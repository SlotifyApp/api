package api

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// parseMSFTAttendees filters out attributes of MSFT attendees.
// see openapi spec to find docs about this.
func parseMSFTAttendees(e graphmodels.Eventable) []Attendee {
	msftAttendees := e.GetAttendees()
	var attendees []Attendee
	attendees = make([]Attendee, 0)

	// Go through MSFT attendees and parse information we need
	for _, a := range msftAttendees {
		var email openapi_types.Email
		if a.GetEmailAddress() != nil && a.GetEmailAddress().GetAddress() != nil {
			emailStr := a.GetEmailAddress().GetAddress()
			email = openapi_types.Email(*emailStr)
		}

		var responseStatus AttendeeResponseStatus
		if a.GetStatus() != nil && a.GetStatus().GetResponse() != nil {
			responseStatus = AttendeeResponseStatus(a.GetStatus().GetResponse().String())
		}

		var attendeeType AttendeeType
		if a.GetTypeEscaped() != nil {
			attendeeType = AttendeeType(a.GetTypeEscaped().String())
		}

		attendee := Attendee{
			Email:          &email,
			ResponseStatus: &responseStatus,
			Type:           &attendeeType,
		}
		attendees = append(attendees, attendee)
	}
	return attendees
}

// parseMSFTLocations filters out attributes of MSFT locations.
// see openapi spec to find docs about this.
func parseMSFTLocations(e graphmodels.Eventable) []Location {
	msftLocations := e.GetLocations()
	var locations []Location
	locations = make([]Location, 0)

	for _, l := range msftLocations {
		var roomType LocationRoomType
		if l.GetLocationType() != nil {
			roomType = LocationRoomType(l.GetLocationType().String())
		}

		var street *string
		if l.GetAddress() != nil {
			street = l.GetAddress().GetStreet()
		}

		parsedLoc := Location{
			Id:       l.GetUniqueId(),
			Name:     l.GetDisplayName(),
			Street:   street,
			RoomType: &roomType,
		}

		locations = append(locations, parsedLoc)
	}
	return locations
}

// parseCalendarEventToMSFTEvent parses CalendarEvent to create a MSFT Event.
func parseCalendarEventToMSFTEvent(eventRequest CalendarEvent) *graphmodels.Event {
	event := graphmodels.NewEvent()
	event.SetSubject(eventRequest.Subject)

	contentType := graphmodels.HTML_BODYTYPE
	body := graphmodels.NewItemBody()
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
	var location *graphmodels.Location
	if len(eventRequest.Locations) > 0 {
		location.SetDisplayName(eventRequest.Locations[0].Name)
	}

	var attendees []graphmodels.Attendeeable
	if eventRequest.Attendees != nil {
		for _, inviteAttendee := range eventRequest.Attendees {
			var email *graphmodels.EmailAddress
			if inviteAttendee.Email != nil {
				email = graphmodels.NewEmailAddress()
				email.SetAddress((*string)(inviteAttendee.Email))
			}

			attendee := graphmodels.NewAttendee()
			attendee.SetEmailAddress(email)

			var attendeeType graphmodels.AttendeeType
			if inviteAttendee.Type != nil {
				switch *inviteAttendee.Type {
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
	}

	event.SetAttendees(attendees)

	transactionID := uuid.New().String()
	event.SetTransactionId(&transactionID)

	return event
}

// parseEventableResp takes in MSFT's version of events and parses them to extract needed attributes.
// See [MSFT Event Properties] for all of the MSFT event properties.
//
// [MSFT Event Properties]: https://learn.microsoft.com/en-us/graph/api/resources/event?view=graph-rest-1.0#properties
func parseEventableResp(events []graphmodels.Eventable) []CalendarEvent {
	calendarEvents := []CalendarEvent{}
	for _, e := range events {
		// Returned interfaces can be nil
		if e == nil {
			continue
		}
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
			Attendees:   attendees,
			Body:        e.GetBodyPreview(),
			Created:     e.GetCreatedDateTime(),
			EndTime:     endTime,
			Id:          e.GetId(),
			IsCancelled: e.GetIsCancelled(),
			JoinURL:     joinURL,
			Locations:   locations,
			Organizer:   (*openapi_types.Email)(e.GetOrganizer().GetEmailAddress().GetAddress()),
			StartTime:   startTime,
			Subject:     e.GetSubject(),
			WebLink:     e.GetWebLink(),
		}
		calendarEvents = append(calendarEvents, ce)
	}
	return calendarEvents
}

// makeCalendarMeAPICall lists a user's events within a certain time range.
// See [MSFT Calendar Me API Call] for docs on the API call made.
//
// [MSFT Calendar Me API Call]:
// https://learn.microsoft.com/en-us/graph/api/calendar-list-calendarview?view=graph-rest-1.0&tabs=http
func makeCalendarMeAPICall(graph *msgraphsdkgo.GraphServiceClient, startTime,
	endTime time.Time,
) ([]CalendarEvent, error) {
	// Prepare request by formatting request parameters correctly.
	start := startTime.Format(time.RFC3339)
	end := endTime.Format(time.RFC3339)

	// default page size is 100, we could iterate through pages but 999 meetings a month
	// should be more than enough
	var pageSize int32 = 999
	requestParameters := &graphusers.ItemCalendarCalendarViewRequestBuilderGetQueryParameters{
		EndDateTime:   &end,
		StartDateTime: &start,
		Top:           &pageSize,
	}

	configuration := &graphusers.ItemCalendarCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	// Make actual API request.

	var events graphmodels.EventCollectionResponseable
	var err error
	err = retry.Do(func() error {
		events, err = graph.Me().Calendar().CalendarView().Get(context.Background(), configuration)
		if err != nil || events == nil {
			return fmt.Errorf("failed to make graph client call: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		return nil, fmt.Errorf("failed msft create event after 3 retries: %w", err)
	}

	// Filter out attributes that we want.

	return parseEventableResp(events.GetValue()), nil
}
