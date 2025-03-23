package api

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	openapi_types "github.com/oapi-codegen/runtime/types"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// parseEventableResp takes in MSFT's version of events and parses them to extract needed attributes.
// See [MSFT Event Properties] for all of the MSFT event properties.
//
// [MSFT Event Properties]: https://learn.microsoft.com/en-us/graph/api/resources/event?view=graph-rest-1.0#properties
func parseEventableResp(events []graphmodels.Eventable) ([]CalendarEvent, error) {
	calendarEvents := []CalendarEvent{}
	for _, e := range events {
		// Returned interfaces can be nil
		if e == nil {
			continue
		}
		attendees, err := parseMSFTAttendees(e)
		if err != nil {
			return nil, fmt.Errorf("failed to parse msft attendees: %w", err)
		}
		locations := parseMSFTLocations(e.GetLocations())

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

		var body *string
		if e.GetBody() != nil {
			body = e.GetBody().GetContent()
		}

		ce := CalendarEvent{
			Attendees:   attendees,
			Body:        body,
			Created:     e.GetCreatedDateTime(),
			EndTime:     endTime,
			Id:          e.GetId(),
			ICalUId:     e.GetICalUId(),
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
	return calendarEvents, nil
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
	var parsedEvents []CalendarEvent
	if parsedEvents, err = parseEventableResp(events.GetValue()); err != nil {
		return nil, fmt.Errorf("failed to parse msft eventable: %w", err)
	}
	return parsedEvents, nil
}
