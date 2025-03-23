package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/avast/retry-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// durationToISO formats a positive duration in the ISO 8601 format.
func durationToISO(duration time.Duration) string {
	hours := int(duration.Hours())
	//nolint: mnd // Magic Number 60
	minutes := int(duration.Minutes()) % 60
	//nolint: mnd // Magic Number 60
	seconds := int(duration.Seconds()) % 60

	isoFormat := "P"
	if hours > 0 || minutes > 0 || seconds > 0 {
		isoFormat += "T"
	}

	if hours > 0 {
		isoFormat += fmt.Sprintf("%dH", hours)
	}
	if minutes > 0 {
		isoFormat += fmt.Sprintf("%dM", minutes)
	}
	if seconds > 0 {
		isoFormat += fmt.Sprintf("%dS", seconds)
	}

	return isoFormat
}

func createSchedulingRequest(body ReschedulingCheckBodySchema,
	meetingPref database.Meetingpreferences,
	msftMeeting graphmodels.Eventable,
) (SchedulingSlotsBodySchema, error) {
	// Create request body for scheduling slots api call
	newReqBody := SchedulingSlotsBodySchema{}

	newReqBody.IsOrganizerOptional = *body.OldMeeting.IsOrganizerOptional
	newReqBody.MeetingName = *msftMeeting.GetSubject()

	minimum := 100.0
	newReqBody.MinimumAttendeePercentage = &minimum
	newReqBody.Attendees = []AttendeeBase{}

	// parse each attendee to attendeeBase
	attendees, err := parseMSFTAttendees(msftMeeting)
	if err != nil {
		return SchedulingSlotsBodySchema{}, fmt.Errorf("failed to parse msft attendees: %w", err)
	}
	for _, a := range attendees {
		ab := AttendeeBase{
			AttendeeType: *a.AttendeeType,
			EmailAddress: EmailAddress{
				Address: a.Email,
				Name:    string(a.Email),
			},
		}
		newReqBody.Attendees = append(newReqBody.Attendees, ab)
	}

	startTime, err := time.Parse(time.RFC3339Nano, *msftMeeting.GetStart().GetDateTime()+"Z")
	if err != nil {
		return newReqBody, fmt.Errorf("failed in parse start time: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339Nano, *msftMeeting.GetEnd().GetDateTime()+"Z")
	if err != nil {
		return newReqBody, fmt.Errorf("failed to parse end time: %w", err)
	}

	if endTime.Before(startTime) {
		return SchedulingSlotsBodySchema{}, errors.New("end time is before the start time, which is invalid")
	}

	duration := endTime.Sub(startTime)
	newReqBody.MeetingDuration = durationToISO(duration)

	// Add time contraints
	timeConstraint := graphmodels.NewTimeConstraint()
	activityDomain := graphmodels.WORK_ACTIVITYDOMAIN
	timeConstraint.SetActivityDomain(&activityDomain)

	timeSlots := []graphmodels.TimeSlotable{}

	timeSlot := graphmodels.NewTimeSlot()
	//nolint: goconst // To make it a constant
	timeZone := "GMT Standard Time"

	startTimeFormatted := meetingPref.StartDateRange.Format(time.RFC3339Nano)
	endTimeFormatted := meetingPref.EndDateRange.Format(time.RFC3339Nano)

	start := graphmodels.NewDateTimeTimeZone()
	start.SetDateTime(&startTimeFormatted)
	start.SetTimeZone(&timeZone)

	end := graphmodels.NewDateTimeTimeZone()
	end.SetDateTime(&endTimeFormatted)
	end.SetTimeZone(&timeZone)

	timeSlot.SetStart(start)
	timeSlot.SetEnd(end)

	timeSlots = append(timeSlots, timeSlot)

	timeConstraint.SetTimeSlots(timeSlots)

	return newReqBody, nil
}

func checkValidReschedulingSlotExists(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body ReschedulingCheckBodySchema,
	msftMeeting graphmodels.Eventable,
	meetingPref database.Meetingpreferences,
) (bool, error) {
	// Call scheduling function to check for valid slots
	newRequest, err := createSchedulingRequest(body, meetingPref, msftMeeting)
	if err != nil {
		return false,
			fmt.Errorf("failed in creating request body for scheduling request: %w", err)
	}

	res, err := makeFindMeetingTimesAPICall(ctx, graph, newRequest)
	if err != nil {
		return false,
			fmt.Errorf("failed in calling find meeting times api: %w", err)
	}

	return len(*res.MeetingTimeSuggestions) > 0, nil
}

func performReschedulingCheckProcess(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body ReschedulingCheckBodySchema,
	msftMeeting graphmodels.Eventable,
	meetingPref database.Meetingpreferences,
) (map[string]bool, error) {
	// Check if the old meeting has valid rescheduling slots

	validSlots, err := checkValidReschedulingSlotExists(ctx, graph, body, msftMeeting, meetingPref)
	if err != nil {
		return nil,
			fmt.Errorf("failed to check valid rescheduling slots exists: %w", err)
	}

	// TODO:
	// Check if the new meeting is more important
	// Simply call AWS Sagemaker AI Endpoint

	response := map[string]bool{
		"isNewMeetingMoreImportant": true,
		"canBeRescheduled":          validSlots,
	}

	return response, nil
}

type NewMeetingAndPrefsParams struct {
	MeetingStartTime time.Time
	OwnerEmail       string
	MsftMeetingID    string
}

func createNewMeetingsAndPrefs(ctx context.Context,
	body NewMeetingAndPrefsParams,
	s Server,
	msftID string,
) (database.Meeting, error) {
	// Meeting Info does not exist so create a new one
	// Check valid user id

	meetingPrefParams := database.CreateMeetingPreferencesParams{
		MeetingStartTime: body.MeetingStartTime,
		StartDateRange:   time.Now(),
		EndDateRange:     body.MeetingStartTime.Add(time.Hour * hoursInAWeek), // 1 week : 24 * 7
	}

	var meetingPrefID int64
	var err error
	err = retry.Do(func() error {
		if meetingPrefID, err = s.DB.CreateMeetingPreferences(ctx, meetingPrefParams); err != nil {
			return fmt.Errorf("failed to create meeting preference: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		return database.Meeting{}, err
	}

	// Create Meeting

	meetingParams := database.CreateMeetingParams{
		//nolint: gosec // id is unsigned 32 bit int
		MeetingPrefID: uint32(meetingPrefID),
		OwnerEmail:    body.OwnerEmail,
		MsftMeetingID: msftID,
	}

	err = retry.Do(func() error {
		if _, err = s.DB.CreateMeeting(ctx, meetingParams); err != nil {
			return fmt.Errorf("failed to create meeting: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		return database.Meeting{}, err
	}

	var meeting database.Meeting
	meeting, err = s.DB.GetMeetingByMSFTID(ctx, msftID)
	if err != nil {
		return database.Meeting{}, err
	}

	return meeting, nil
}

func processNewMeetingInfo(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	s Server,
	msftMeetingID string,
	userEmail string,
) (database.Meeting, error) {
	// Fetch meeting data from microsft

	msftMeeting, err := getUsersEvent(ctx, graph, msftMeetingID)
	if err != nil {
		return database.Meeting{}, fmt.Errorf("failed to get meeting data from microsoft: %w", err)
	}

	var startTime time.Time
	startTime, err = time.Parse(time.RFC3339Nano, *msftMeeting.GetStart().GetDateTime()+"Z")
	if err != nil {
		return database.Meeting{}, fmt.Errorf("failed to get parse start time: %w", err)
	}

	newMeetingParams := NewMeetingAndPrefsParams{
		MeetingStartTime: startTime,
		OwnerEmail:       userEmail,
		MsftMeetingID:    msftMeetingID,
	}

	meeting, err := createNewMeetingsAndPrefs(ctx, newMeetingParams, s, msftMeetingID)
	if err != nil {
		return database.Meeting{}, fmt.Errorf("failed to get data from new db.Meeting: %w", err)
	}

	return meeting, nil
}

func getUsersEvent(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	msftID string,
) (graphmodels.Eventable, error) {
	queryFilter := "iCalUId eq '" + msftID + "'"

	// Get old meeting data from microsoft
	requestConfig := users.ItemEventsRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemEventsRequestBuilderGetQueryParameters{
			Filter: &queryFilter,
		},
	}

	msftMeetingRes, err := graph.Me().Events().Get(ctx, &requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get meeting data from microsoft: %w", err)
	}

	if msftMeetingRes != nil && msftMeetingRes.GetValue() != nil &&
		len(msftMeetingRes.GetValue()) > 0 {
		return msftMeetingRes.GetValue()[0], nil
	}

	return nil, errors.New("failed to get meeting data from microsoft: returned empty array")
}
