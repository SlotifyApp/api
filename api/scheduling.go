package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/avast/retry-go"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/microsoft/kiota-abstractions-go/serialization"
)

type SchedulingGraphReq struct {
	reqBody *graphusers.ItemFindMeetingTimesPostRequestBody
	config  *graphusers.ItemFindMeetingTimesRequestBuilderPostRequestConfiguration
}

// ProcessFindMeetingsResponse process the findMeeting graph response into the response body.
func ProcessFindMeetingsResponse(res graphmodels.MeetingTimeSuggestionsResultable) (
	SchedulingSlotsSuccessResponseBody, error,
) {
	if res == nil {
		return SchedulingSlotsSuccessResponseBody{}, errors.New("cannot process an empty MeetingTimeSuggestionResultable")
	}

	resp := SchedulingSlotsSuccessResponseBody{}

	emptySuggestionReason := res.GetEmptySuggestionsReason()
	if emptySuggestionReason != nil {
		reason := getEmptySuggestionsReason(*emptySuggestionReason)
		resp.EmptySuggestionsReason = &reason
	}

	suggestions := res.GetMeetingTimeSuggestions()

	processedMeetingSuggestions := make([]MeetingTimeSuggestion, 0)
	for _, s := range suggestions {
		if s == nil {
			continue
		}
		m := MeetingTimeSuggestion{
			Confidence:       s.GetConfidence(),
			Order:            s.GetOrder(),
			SuggestionReason: s.GetSuggestionReason(),
		}

		// Process organiser availability
		if organiserAvailibility := s.GetOrganizerAvailability(); organiserAvailibility != nil {
			oa := organiserAvailibility.String()
			m.OrganizerAvailability = &oa
		}

		// Process time slot
		if timeSlot := processMSFTTimeSlot(s); timeSlot != nil {
			m.MeetingTimeSlot = timeSlot
		}

		processedLocs := parseMSFTLocations(s.GetLocations())
		m.Locations = &processedLocs

		processedAttendeeAvailabilities := processMSFTAttendeeAvailabilities(s)
		m.AttendeeAvailability = &processedAttendeeAvailabilities

		processedMeetingSuggestions = append(processedMeetingSuggestions, m)
	}

	resp.MeetingTimeSuggestions = &processedMeetingSuggestions

	return resp, nil
}

// processReqBodyTimeConstraints will transform time constraints into msgraph TimeSlotable.
func processReqBodyTimeConstraints(body *SchedulingSlotsBodySchema) *graphmodels.TimeConstraint {
	// TODO: See if user wants to respsect user working hours and then set time constraint.
	// See recent meeting notes
	timeConstraint := graphmodels.NewTimeConstraint()
	// TODO: get from req body
	activityDomain := graphmodels.WORK_ACTIVITYDOMAIN
	timeConstraint.SetActivityDomain(&activityDomain)

	timeSlots := []graphmodels.TimeSlotable{}
	for _, ts := range body.TimeConstraint.TimeSlots {
		timeSlot := graphmodels.NewTimeSlot()
		timeZone := "GMT Standard Time"

		startTimeFormatted := ts.Start.Format(time.RFC3339Nano)
		endTimeFormatted := ts.End.Format(time.RFC3339Nano)

		start := graphmodels.NewDateTimeTimeZone()
		start.SetDateTime(&startTimeFormatted)
		start.SetTimeZone(&timeZone)

		end := graphmodels.NewDateTimeTimeZone()
		end.SetDateTime(&endTimeFormatted)
		end.SetTimeZone(&timeZone)

		timeSlot.SetStart(start)
		timeSlot.SetEnd(end)

		timeSlots = append(timeSlots, timeSlot)
	}
	timeConstraint.SetTimeSlots(timeSlots)

	return timeConstraint
}

// processReqBodyAttendees processes the scheduling slots request body for a msgraph Attendee list.
func processReqBodyAttendees(body *SchedulingSlotsBodySchema) []graphmodels.AttendeeBaseable {
	attendees := []graphmodels.AttendeeBaseable{}
	for _, a := range body.Attendees {
		attendeeBase := graphmodels.NewAttendeeBase()

		emailAddress := graphmodels.NewEmailAddress()
		emailAddress.SetName(&a.EmailAddress.Name)
		emailAddress.SetAddress((*string)(&a.EmailAddress.Address))

		attendeeBase.SetEmailAddress(emailAddress)

		attendeeType := graphmodels.REQUIRED_ATTENDEETYPE
		switch a.AttendeeType {
		case Optional:
			attendeeType = graphmodels.OPTIONAL_ATTENDEETYPE
		case Resource:
			attendeeType = graphmodels.RESOURCE_ATTENDEETYPE
		case Required:
			attendeeType = graphmodels.REQUIRED_ATTENDEETYPE
		}

		attendeeBase.SetTypeEscaped(&attendeeType)

		attendees = append(attendees, attendeeBase)
	}

	return attendees
}

// processReqBodyLocationConstraint processes the scheduling slots request body for a msgraph LocationConstraint.
func processReqBodyLocationConstraint(body *SchedulingSlotsBodySchema) *graphmodels.LocationConstraint {
	locationConstraint := graphmodels.NewLocationConstraint()
	locations := []graphmodels.LocationConstraintItemable{}
	if body.LocationConstraint.Locations != nil {
		for _, lc := range *body.LocationConstraint.Locations {
			locationConstraintItem := graphmodels.NewLocationConstraintItem()
			locationConstraintItem.SetResolveAvailability(&lc.ResolveAvailability)
			locationConstraintItem.SetDisplayName(&lc.DisplayName)
			locationConstraintItem.SetLocationEmailAddress(lc.LocationEmailAddress)

			physicalAddr := graphmodels.NewPhysicalAddress()
			physicalAddr.SetCity(lc.Address.City)
			physicalAddr.SetPostalCode(lc.Address.PostalCode)
			physicalAddr.SetStreet(lc.Address.Street)
			physicalAddr.SetState(lc.Address.State)
			physicalAddr.SetCountryOrRegion(lc.Address.CountryOrRegion)

			locationConstraintItem.SetAddress(physicalAddr)

			locations = append(locations, locationConstraintItem)
		}
	}
	locationConstraint.SetLocations(locations)

	locationConstraint.SetIsRequired(body.LocationConstraint.IsRequired)
	locationConstraint.SetSuggestLocation(body.LocationConstraint.SuggestLocation)

	return locationConstraint
}

// CreateSchedulingGraphReqBody will create the header and body for the findMeeting graph endpoint.
func CreateSchedulingGraphReqBody(body *SchedulingSlotsBodySchema) (SchedulingGraphReq, error) {
	// Create graph headers
	headers := abstractions.NewRequestHeaders()
	headers.Add("Prefer", "outlook.timezone=\"GMT Standard Time\"")

	configuration := &graphusers.ItemFindMeetingTimesRequestBuilderPostRequestConfiguration{
		Headers: headers,
	}

	// Create graph request body
	graphRequestBody := graphusers.NewItemFindMeetingTimesPostRequestBody()

	// Set attendees
	attendees := processReqBodyAttendees(body)
	graphRequestBody.SetAttendees(attendees)

	// Set Locations
	lc := processReqBodyLocationConstraint(body)
	graphRequestBody.SetLocationConstraint(lc)

	// Set time constraints
	tc := processReqBodyTimeConstraints(body)
	graphRequestBody.SetTimeConstraint(tc)

	// Set meeting duration
	meetingDuration, err := serialization.ParseISODuration(body.MeetingDuration)
	if err != nil {
		return SchedulingGraphReq{},
			fmt.Errorf("failed to parse meeting duration %s as ISO Duration: %w", body.MeetingDuration, err)
	}

	graphRequestBody.SetMeetingDuration(meetingDuration)

	// Set optional params
	graphRequestBody.SetIsOrganizerOptional(&body.IsOrganizerOptional)
	returnSuggestionReasons := true
	graphRequestBody.SetReturnSuggestionReasons(&returnSuggestionReasons)

	graphRequestBody.SetMinimumAttendeePercentage(body.MinimumAttendeePercentage)

	if body.MaxCandidates != nil {
		graphRequestBody.SetMaxCandidates(body.MaxCandidates)
	} else {
		var defaultMaxCandidates int32 = 10
		graphRequestBody.SetMaxCandidates(&defaultMaxCandidates)
	}

	return SchedulingGraphReq{
		config:  configuration,
		reqBody: graphRequestBody,
	}, nil
}

func makeFindMeetingTimesAPICall(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body SchedulingSlotsBodySchema) (
	SchedulingSlotsSuccessResponseBody, error,
) {
	// Get custom graph request header and config
	graphConfigAndBody, err := CreateSchedulingGraphReqBody(&body)
	if err != nil {
		return SchedulingSlotsSuccessResponseBody{},
			fmt.Errorf("failed to create graph req body for findMeetings: %w", err)
	}

	// Attempt to call FindMeetingTimes 3 times
	var findMeetingTimes graphmodels.MeetingTimeSuggestionsResultable
	err = retry.Do(func() error {
		findMeetingTimes, err = graph.Me().FindMeetingTimes().Post(ctx, graphConfigAndBody.reqBody, graphConfigAndBody.config)
		if err != nil {
			return fmt.Errorf("failed to make msgraph findMeeting API call: %w", err)
		}

		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		return SchedulingSlotsSuccessResponseBody{},
			fmt.Errorf("failed msft find meeting times after 3 retries: %w", err)
	}

	// Process MSFT resp into our own types
	respBody, err := ProcessFindMeetingsResponse(findMeetingTimes)
	if err != nil {
		return SchedulingSlotsSuccessResponseBody{},
			fmt.Errorf("failed to process msft findMeeting response: %w", err)
	}

	return respBody, nil
}

func getUserWorkingHours(s Server,
	calendarEvent []CalendarEvent,
	userWorkingHours map[string]float64,
) map[string]float64 {
	// Calculate working hours for each day
	for _, event := range calendarEvent {
		startTime, err := time.Parse(time.RFC3339, *event.StartTime+"Z")
		if err != nil {
			s.Logger.Error("failed to parse calendar start time", zap.Error(err))
			continue
		}

		var endTime time.Time
		endTime, err = time.Parse(time.RFC3339, *event.EndTime+"Z")
		if err != nil {
			s.Logger.Error("failed to parse calendar end time", zap.Error(err))
			continue
		}

		duration := endTime.Sub(startTime).Minutes()

		val, ok := userWorkingHours[startTime.Format("2006-01-02")]

		if !ok {
			userWorkingHours[startTime.Format("2006-01-02")] = duration
		} else {
			userWorkingHours[startTime.Format("2006-01-02")] = val + duration
		}
	}

	return userWorkingHours
}

func generateRatingsForSlots(ctx context.Context,
	s Server,
	ownerID uint32,
	possibleSlots SchedulingSlotsSuccessResponseBody,
	body SchedulingSlotsBodySchema,
) SchedulingSlotsSuccessResponseBody {
	// For each attendee, get their calendar data to calculate how many hours of meetings they have each day
	// then update the running average
	const workingMinutes = float64(540) // 9 hours * 60 minutes
	response := possibleSlots

	tempBody := body

	ownerUser, err := s.DB.GetUserByID(ctx, ownerID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
	} else {
		tempBody.Attendees = append(tempBody.Attendees, AttendeeBase{
			AttendeeType: "required",
			EmailAddress: EmailAddress{
				Address: types.Email(ownerUser.Email),
				Name:    ownerUser.FirstName,
			},
		})
	}

	proAttende := processReqBodyAttendees(&tempBody)

	for no, a := range proAttende {
		var aEmailAdd string
		if a.GetEmailAddress() != nil || a.GetEmailAddress().GetAddress() != nil {
			aEmailAdd = *a.GetEmailAddress().GetAddress()
		} else {
			continue
		}

		// Get user ID
		var user database.User
		user, err = s.DB.GetUserByEmail(ctx, aEmailAdd)
		if err != nil {
			s.Logger.Error("failed to create msgraph client", zap.Error(err))
			continue
		}

		// Fetch calendar for user
		var graph *msgraphsdkgo.GraphServiceClient
		graph, err = CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, user.ID)
		if err != nil {
			s.Logger.Error("failed to create msgraph client", zap.Error(err))
			continue
		}
		userWorkingHours := map[string]float64{}

		// Make call to API route and parse events
		var calendarEvent []CalendarEvent
		calendarEvent, err = makeCalendarMeAPICall(graph,
			body.TimeConstraint.TimeSlots[0].Start,
			body.TimeConstraint.TimeSlots[0].End)
		if err != nil {
			s.Logger.Error("failed to make calendar me msgraph api call", zap.Error(err))
			continue
		}

		userWorkingHours = getUserWorkingHours(s, calendarEvent, userWorkingHours)

		// Update running average of confidence
		for idx, slot := range *possibleSlots.MeetingTimeSuggestions {
			dur := slot.MeetingTimeSlot.End.Sub(slot.MeetingTimeSlot.End)
			totalTime, ok := userWorkingHours[slot.MeetingTimeSlot.Start.Format("2006-01-02")]

			if ok {
				totalTime = 0
			}

			remainingMinutes := workingMinutes - dur.Minutes() - totalTime
			var personalScore float64
			if remainingMinutes > 0 {
				//nolint: mnd // magic number 100 for converting to percentage from decimal
				personalScore = (remainingMinutes / workingMinutes) * 100
			} else {
				personalScore = 0
			}

			if no == 0 {
				(*response.MeetingTimeSuggestions)[idx].Confidence = &personalScore
			} else {
				// running average
				curr := *(*response.MeetingTimeSuggestions)[idx].Confidence * float64(no)
				updatedConfidence := (personalScore + curr) / float64(no+1)
				(*response.MeetingTimeSuggestions)[idx].Confidence = &updatedConfidence
			}
		}
	}

	return response
}
