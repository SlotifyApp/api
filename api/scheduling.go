package api

import (
	abstractions "github.com/microsoft/kiota-abstractions-go"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/microsoft/kiota-abstractions-go/serialization"
)

type SchedulingGraphReq struct {
	reqBody *graphusers.ItemFindMeetingTimesPostRequestBody
	config  *graphusers.ItemFindMeetingTimesRequestBuilderPostRequestConfiguration
}

func CreateSchedulingGraphReqBody() (SchedulingGraphReq, error) {
	// Create graph headers
	headers := abstractions.NewRequestHeaders()
	headers.Add("Prefer", "outlook.timezone=\"GMT Standard Time\"")

	configuration := &graphusers.ItemFindMeetingTimesRequestBuilderPostRequestConfiguration{
		Headers: headers,
	}

	// Create graph request body
	graphRequestBody := graphusers.NewItemFindMeetingTimesPostRequestBody()

	// Set attendees
	attendeeBase := graphmodels.NewAttendeeBase()
	attendeeType := graphmodels.REQUIRED_ATTENDEETYPE
	attendeeBase.SetTypeEscaped(&attendeeType)
	emailAddress := graphmodels.NewEmailAddress()
	name := "Alex Wilbur"
	emailAddress.SetName(&name)
	address := "alexw@contoso.com"
	emailAddress.SetAddress(&address)
	attendeeBase.SetEmailAddress(emailAddress)

	attendees := []graphmodels.AttendeeBaseable{
		attendeeBase,
	}
	graphRequestBody.SetAttendees(attendees)

	// Set Locations
	locationConstraint := graphmodels.NewLocationConstraint()
	isRequired := false
	locationConstraint.SetIsRequired(&isRequired)
	suggestLocation := false
	locationConstraint.SetSuggestLocation(&suggestLocation)

	locationConstraintItem := graphmodels.NewLocationConstraintItem()
	resolveAvailability := false
	locationConstraintItem.SetResolveAvailability(&resolveAvailability)
	displayName := "Conf room Hood"
	locationConstraintItem.SetDisplayName(&displayName)

	locations := []graphmodels.LocationConstraintItemable{
		locationConstraintItem,
	}
	locationConstraint.SetLocations(locations)
	graphRequestBody.SetLocationConstraint(locationConstraint)
	timeConstraint := graphmodels.NewTimeConstraint()
	activityDomain := graphmodels.WORK_ACTIVITYDOMAIN
	timeConstraint.SetActivityDomain(&activityDomain)

	timeSlot := graphmodels.NewTimeSlot()
	start := graphmodels.NewDateTimeTimeZone()
	dateTime := "2019-04-16T09:00:00"
	start.SetDateTime(&dateTime)
	timeZone := "Pacific Standard Time"
	start.SetTimeZone(&timeZone)
	timeSlot.SetStart(start)
	end := graphmodels.NewDateTimeTimeZone()
	dateTime = "2019-04-18T17:00:00"
	end.SetDateTime(&dateTime)
	timeZone = "Pacific Standard Time"
	end.SetTimeZone(&timeZone)
	timeSlot.SetEnd(end)

	timeSlots := []graphmodels.TimeSlotable{
		timeSlot,
	}
	timeConstraint.SetTimeSlots(timeSlots)
	graphRequestBody.SetTimeConstraint(timeConstraint)
	isOrganizerOptional := false
	graphRequestBody.SetIsOrganizerOptional(&isOrganizerOptional)

	meetingDuration, err := serialization.ParseISODuration("PT1H")
	if err != nil {
		// TODO: Wrap this
		return SchedulingGraphReq{}, err
	}

	graphRequestBody.SetMeetingDuration(meetingDuration)
	returnSuggestionReasons := true
	graphRequestBody.SetReturnSuggestionReasons(&returnSuggestionReasons)
	minimumAttendeePercentage := float64(100)
	graphRequestBody.SetMinimumAttendeePercentage(&minimumAttendeePercentage)

	return SchedulingGraphReq{
		config:  configuration,
		reqBody: graphRequestBody,
	}, nil
}
