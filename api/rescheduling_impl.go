package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

// (POST /api/reschedule/check).
func (s Server) PostAPIRescheduleCheck(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body ReschedulingCheckBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		s.Logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Get data from db
	var meeting database.Meeting
	meeting, err = s.DB.GetMeetingByID(ctx, uint32(*body.OldMeeting.MeetingID))

	if err != nil {
		s.Logger.Error("failed to search meeting table in db", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process db request")
		return
	}

	// TODO: Add check for if no meeting is found

	var meetingPref database.Meetingpreferences
	meetingPref, err = s.DB.GetMeetingPreferences(ctx, meeting.MeetingPrefID)

	if err != nil {
		s.Logger.Error("failed to search meeting table in db", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process db request")
		return
	}

	respBody, err := performReschedulingCheckProcess(ctx, graph, body, meetingPref)
	if err != nil {
		s.Logger.Error("failed to make msgraph api call to findMeetings", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process/send microsoft graph API request for findMeeting")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, respBody)
}

// (POST /api/reschedule/request).
func (s Server) PostAPIRescheduleRequest(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body ReschedulingRequestBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		s.Logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	// graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	// if err != nil {
	// 	s.Logger.Error("failed to create msgraph client", zap.Error(err))
	// 	sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
	// 	return
	// }

	// Get data from db
	var meeting database.Meeting
	meeting, err = s.DB.GetMeetingByID(ctx, uint32(*body.OldMeeting.MeetingID))

	if err != nil {
		s.Logger.Error("failed to get data from db.Meeting", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get data from db.Meeting")
		return
	}

	// Create Rescheduling Request
	var requestId int64
	err = retry.Do(func() error {
		if requestId, err = s.DB.CreateReschedulingRequest(ctx, meeting.ID); err != nil {
			return fmt.Errorf("Failed to create reschedling requested by user: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))

	//Create requested by time
	params := database.CreateReschedulingRequestedByUserParams{
		RequestID: uint32(requestId),
		UserID:    userID,
	}

	err = retry.Do(func() error {
		if _, err = s.DB.CreateReschedulingRequestedByUser(ctx, params); err != nil {
			return fmt.Errorf("Failed to create reschedling requested by user: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))

	// Create Placeholder meeting info
	placeholderParams := database.CreatePlaceholderMeetingParams{
		RequestID:      uint32(requestId),
		Title:          *body.NewMeeting.Title,
		OwnerID:        userID,
		StartTime:      *body.NewMeeting.StartTime,
		EndTime:        *body.NewMeeting.EndTime,
		Location:       *body.NewMeeting.Location,
		Duration:       int32(*body.NewMeeting.Duration),
		StartDateRange: *body.NewMeeting.StartRangeTime,
		EndDateRange:   *body.NewMeeting.EndRangeTime,
	}

	var placeholderMeeting int64
	err = retry.Do(func() error {
		if placeholderMeeting, err = s.DB.CreatePlaceholderMeeting(ctx, placeholderParams); err != nil {
			return fmt.Errorf("Failed to create placeholder meeting: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))

	// For each attendee, create placeholder attendee row
	for _, attendee := range *body.NewMeeting.Attendees {
		attendeeParams := database.CreatePlaceholderMeetingAttendeeParams{
			MeetingID: int32(placeholderMeeting),
			UserID:    uint32(len(attendee.EmailAddress.Address)), // TODO: Fix this
		}

		err = retry.Do(func() error {
			if _, err = s.DB.CreatePlaceholderMeetingAttendee(ctx, attendeeParams); err != nil {
				return fmt.Errorf("Failed to create placeholder meeting attendee: %w", err)
			}
			return nil
		}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	}

	// NOtify user of the request

	if err != nil {
		s.Logger.Error("failed to create request", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create request")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Created")
}
