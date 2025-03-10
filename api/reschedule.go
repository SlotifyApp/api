package api

import (
	"context"
	"fmt"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

func createSchedulingRequest(body ReschedulingCheckBodySchema,
	meetingPref database.Meetingpreferences,
) (SchedulingSlotsBodySchema, error) {
	// Create request body for scheduling slots api call
	newReqBody := SchedulingSlotsBodySchema{}

	newReqBody.Attendees = *body.OldMeeting.Attendees
	newReqBody.IsOrganizerOptional = *body.OldMeeting.IsOrganizerOptional
	newReqBody.MeetingName = *body.OldMeeting.Title
	newReqBody.MeetingDuration = *body.OldMeeting.MeetingDuration

	minimum := 100.0
	newReqBody.MinimumAttendeePercentage = &minimum
	newReqBody.LocationConstraint = *body.OldMeeting.LocationConstraint

	// Add time contraints
	timeConstraint := graphmodels.NewTimeConstraint()
	activityDomain := graphmodels.WORK_ACTIVITYDOMAIN
	timeConstraint.SetActivityDomain(&activityDomain)

	timeSlots := []graphmodels.TimeSlotable{}

	timeSlot := graphmodels.NewTimeSlot()
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
	meetingPref database.Meetingpreferences,
) (bool, error) {
	// Call scheduling function to check for valid slots
	newRequest, err := createSchedulingRequest(body, meetingPref)
	if err != nil {
		return false,
			fmt.Errorf("failed in creating find meeting time request body: %w", err)
	}

	res, err := makeFindMeetingTimesAPICall(ctx, graph, newRequest)
	if err != nil {
		return false,
			fmt.Errorf("failed in calling fine meeting times api: %w", err)
	}

	// Check if valid slots exist
	if len(*res.MeetingTimeSuggestions) > 0 {
		return true, nil
	}

	return false, nil
}

func performReschedulingCheckProcess(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body ReschedulingCheckBodySchema,
	meetingPref database.Meetingpreferences,
) (map[string]bool, error) {
	// Check if the old meeting has valid rescheduling slots

	validSlots, err := checkValidReschedulingSlotExists(ctx, graph, body, meetingPref)
	if err != nil {
		return nil,
			fmt.Errorf("failed to check valid rescheduling slots exists: %w", err)
	}

	// TODO
	// Check if the new meeting is more important
	// Simply call AWS Sagemaker AI Endpoint

	response := map[string]bool{
		"isNewMeetingMoreImportant": true,
		"canBeRescheduled":          validSlots,
	}

	return response, nil
}
