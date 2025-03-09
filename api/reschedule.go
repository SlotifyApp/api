package api

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func createSchedulingRequest(body ReschedulingCheckBodySchema) (SchedulingSlotsBodySchema, error) {
	// Create request body for scheduling slots api call
	newReqBody := SchedulingSlotsBodySchema{}

	newReqBody.Attendees = *body.OldMeeting.Attendees
	newReqBody.IsOrganizerOptional = *body.OldMeeting.IsOrganizerOptional
	newReqBody.MeetingName = *body.OldMeeting.Title
	newReqBody.MeetingDuration = *body.OldMeeting.MeetingDuration

	minimum := 100.0
	newReqBody.MinimumAttendeePercentage = &minimum
	newReqBody.LocationConstraint = *body.OldMeeting.LocationConstraint

	// TODO
	// For the time constraints, check a week before and after

	return newReqBody, nil
}

func checkValidReschedulingSlotExists(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body ReschedulingCheckBodySchema,
) (bool, error) {
	// Call scheduling function to check for valid slots
	newRequest, err := createSchedulingRequest(body)
	if err != nil {
		return false,
			fmt.Errorf("failed in creating find meeting time request body: %w", err)
	}

	res, err := makeFindMeetingTimesAPICall(ctx, graph, newRequest)
	if err != nil {
		return false,
			fmt.Errorf("failed in calling fine meeting times api: %w", err)
	}

	if len(*res.MeetingTimeSuggestions) > 0 {
		return true, nil
	}
	// Check if res has non-empty array
	return false, nil
}

func performReschedulingProcess(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body ReschedulingCheckBodySchema,
) (map[string]bool, error) {
	// Check if the old meeting has valid rescheduling slots

	validSlots, err := checkValidReschedulingSlotExists(ctx, graph, body)
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

func processReschedulingRequest(ctx context.Context,
	graph *msgraphsdkgo.GraphServiceClient,
	body ReschedulingRequestBodySchema,
) (string, error) {
	// Create placeholder meeting location

	// Add data into the database

	return "Request failed to be created", nil
}
