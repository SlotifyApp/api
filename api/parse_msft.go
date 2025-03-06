package api

import (
	"time"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"
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
			AttendeeType:   &attendeeType,
		}
		attendees = append(attendees, attendee)
	}
	return attendees
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
			if inviteAttendee.AttendeeType != nil {
				switch *inviteAttendee.AttendeeType {
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

// getFreeBusyStatus attempts to convert a string into a AttendeeType, if none match use FreeBusyStatusUnknown.
func getFreeBusyStatus(s string) FreeBusyStatus {
	switch s {
	case string(FreeBusyStatusBusy):
		return FreeBusyStatusBusy
	case string(FreeBusyStatusFree):
		return FreeBusyStatusFree
	case string(FreeBusyStatusOof):
		return FreeBusyStatusOof
	case string(FreeBusyStatusTentative):
		return FreeBusyStatusTentative
	case string(FreeBusyStatusWorkingElsewhere):
		return FreeBusyStatusWorkingElsewhere
	default:
		return FreeBusyStatusUnknown
	}
}

// getAttendeeType attempts to convert a string into a AttendeeType, if none match use Optional.
func getAttendeeType(s string) AttendeeType {
	switch s {
	case string(Required):
		return Required
	case string(Resource):
		return Resource
	default:
		return Optional
	}
}

// getEmptySuggestionsReason attempts to convert a string into a EmptySuggestionsReason, uses a default type.
func getEmptySuggestionsReason(s string) EmptySuggestionsReason {
	switch s {
	case string(EmptySuggestionsReasonAttendeesUnavailable):
		return EmptySuggestionsReasonAttendeesUnavailable

	case string(EmptySuggestionsReasonAttendeesUnavailableOrUnknown):
		return EmptySuggestionsReasonAttendeesUnavailableOrUnknown

	case string(EmptySuggestionsReasonLocationsUnavailable):
		return EmptySuggestionsReasonLocationsUnavailable

	case string(EmptySuggestionsReasonOrganizerUnavailable):
		return EmptySuggestionsReasonOrganizerUnavailable

	default:
		return EmptySuggestionsReasonAttendeesUnavailable
	}
}

// processMSFTTimeSlot process a msft time slot.
func processMSFTTimeSlot(s graphmodels.MeetingTimeSuggestionable) *MeetingTimeSlot {
	timeSlot := s.GetMeetingTimeSlot()
	if timeSlot == nil {
		return nil
	}
	processedMeetingTimeSlot := MeetingTimeSlot{}
	start := timeSlot.GetStart()
	end := timeSlot.GetEnd()

	layout := "2006-01-02T15:04:05.0000000"

	if start != nil && start.GetDateTime() != nil {
		if t, err := time.Parse(layout, *start.GetDateTime()); err == nil {
			processedMeetingTimeSlot.Start = t
		}
	}

	if end != nil && end.GetDateTime() != nil {
		if t, err := time.Parse(layout, *end.GetDateTime()); err == nil {
			processedMeetingTimeSlot.End = t
		}
	}

	return &processedMeetingTimeSlot
}

// processMSFTAttendeeAvailabilities processes the msft attendee availability.
func processMSFTAttendeeAvailabilities(s graphmodels.MeetingTimeSuggestionable) []AttendeeAvailability {
	attendeeAvailabilities := s.GetAttendeeAvailability()
	processedAttendeeAvailabilities := make([]AttendeeAvailability, 0)
	for _, a := range attendeeAvailabilities {
		if a.GetAttendee() == nil {
			continue
		}

		processedEmailAddress := EmailAddress{}
		emailAddress := a.GetAttendee().GetEmailAddress()

		if emailAddress != nil {
			name := emailAddress.GetName()
			address := emailAddress.GetAddress()
			if name != nil {
				processedEmailAddress.Name = *name
			}
			if address != nil {
				processedEmailAddress.Address = openapi_types.Email(*address)
			}
		}

		attendeeBase := AttendeeBase{
			EmailAddress: processedEmailAddress,
		}

		// set AttendeeType
		attendeeType := a.GetAttendee().GetTypeEscaped()
		if attendeeType != nil {
			attendeeBase.AttendeeType = getAttendeeType(attendeeType.String())
		}

		processedAttendeeAvailability := AttendeeAvailability{
			Attendee:     attendeeBase,
			Availability: FreeBusyStatusUnknown,
		}

		// set AttendeeType
		attendeeAvailability := a.GetAvailability()
		if attendeeAvailability != nil {
			processedAttendeeAvailability.Availability = getFreeBusyStatus(attendeeAvailability.String())
		}

		processedAttendeeAvailabilities = append(processedAttendeeAvailabilities,
			processedAttendeeAvailability)
	}

	return processedAttendeeAvailabilities
}

// getRoomType attempts to convert a string into a LocationRoomType- uses the Default room type.
func getRoomType(r string) LocationRoomType {
	switch r {
	case string(BusinessAddress):
		return BusinessAddress

	case string(ConferenceRoom):
		return ConferenceRoom

	case string(GeoCoordinates):
		return GeoCoordinates

	case string(HomeAddress):
		return HomeAddress

	case string(Hotel):
		return Hotel

	case string(LocalBusiness):
		return LocalBusiness

	case string(PostalAddress):
		return PostalAddress

	case string(Restaurant):
		return Restaurant

	case string(StreetAddress):
		return StreetAddress

	default:
		return Default
	}
}

// parseMSFTLocations filters out attributes of MSFT locations.
// see openapi spec to find docs about this.
func parseMSFTLocations(msftLocations []graphmodels.Locationable) []Location {
	var locations []Location
	locations = make([]Location, 0)

	for _, l := range msftLocations {
		var roomType LocationRoomType
		if l.GetLocationType() != nil {
			roomType = getRoomType(l.GetLocationType().String())
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
