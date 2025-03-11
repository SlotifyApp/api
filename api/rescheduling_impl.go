package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/avast/retry-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

const hoursInAWeek = 168

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
	//nolint: gosec // id is unsigned 32 bit int
	meeting, err = s.DB.GetMeetingByID(ctx, uint32(*body.OldMeeting.MeetingID))

	var meetingFound bool

	if errors.Is(err, sql.ErrNoRows) {
		// Meeting Not Found
		meetingFound = false
	} else if err != nil {
		s.Logger.Error("failed to search meeting table in db", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process db request")
		return
	}

	var meetingPref database.Meetingpreferences
	if meetingFound {
		// Get meeting preferences if data exists
		meetingPref, err = s.DB.GetMeetingPreferences(ctx, meeting.MeetingPrefID)
		if err != nil {
			s.Logger.Error("failed to search meeting table in db", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to process db request")
			return
		}
	} else {
		// Create temp meeting preferences if data doesn't exist

		dayTime := time.Hour * hoursInAWeek // 1 week : 24 * 7
		meetingPref = database.Meetingpreferences{
			StartDateRange: time.Now(),
			EndDateRange:   body.OldMeeting.StartTime.Add(dayTime), // Give a week extra from the start of the meeting
		}
	}

	respBody, err := performReschedulingCheckProcess(ctx, graph, body, meetingPref)
	if err != nil {
		s.Logger.Error("failed to make msgraph api call to findMeetings", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process/send microsoft graph API request for findMeeting")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, respBody)
}

// (POST /api/reschedule/request/replace).
// nolint: funlen // 20 lines too long
func (s Server) PostAPIRescheduleRequestReplace(w http.ResponseWriter, r *http.Request) {
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

	// Create Rescheduling Request
	var requestID int64
	err = retry.Do(func() error {
		if requestID, err = s.DB.CreateReschedulingRequest(ctx, userID); err != nil {
			return fmt.Errorf("failed to create reschedling requested by user: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))

	// Attach meeting preferences info to request if it exists

	// Link request to old meeting
	var meeting database.Meeting
	// Get data from db to validate meeting id

	meeting, err = s.DB.GetMeetingByMSFTID(ctx, *body.OldMeeting.MsftMeetingID)

	if errors.Is(err, sql.ErrNoRows) {
		newMeetingParams := NewMeetingAndPrefsParams{
			MeetingStartTime: *body.OldMeeting.MeetingStartTime,
			//nolint: gosec // id is unsigned 32 bit int
			OwnerID:       uint32(*body.OldMeeting.MeetingOwner),
			MsftMeetingID: *body.OldMeeting.MsftMeetingID,
		}

		meeting, err = createNewMeetingsAndPrefs(ctx, newMeetingParams, s)
		if err != nil {
			s.Logger.Error("failed to get data from new db.Meeting", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get data from new db.Meeting")
			return
		}
	} else if err != nil {
		s.Logger.Error("failed to get data from db.Meeting", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get data from db.Meeting")
		return
	}

	// Create request to meeting
	requestToMeetingParams := database.CreateRequestToMeetingParams{
		//nolint: gosec // id is unsigned 32 bit int
		RequestID: uint32(requestID),
		MeetingID: meeting.ID,
	}

	err = retry.Do(func() error {
		if _, err = s.DB.CreateRequestToMeeting(ctx, requestToMeetingParams); err != nil {
			return fmt.Errorf("failed to create request to meeting link: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("DB Creation Error: ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create request to meeting link")
		return
	}

	// Create Placeholder meeting info
	parsedTime, err := time.Parse(time.RFC3339Nano, *body.NewMeeting.MeetingDuration)
	if err != nil {
		s.Logger.Error("failed to parse time", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to parse time")
		return
	}

	placeholderParams := database.CreatePlaceholderMeetingParams{
		//nolint: gosec // id is unsigned 32 bit int
		RequestID: uint32(requestID),
		Title:     *body.NewMeeting.Title,
		StartTime: *body.NewMeeting.StartTime,
		EndTime:   *body.NewMeeting.EndTime,
		Location:  *body.NewMeeting.Location,

		Duration:       parsedTime,
		StartDateRange: *body.NewMeeting.StartRangeTime,
		EndDateRange:   *body.NewMeeting.EndRangeTime,
	}

	var placeholderMeeting int64
	err = retry.Do(func() error {
		if placeholderMeeting, err = s.DB.CreatePlaceholderMeeting(ctx, placeholderParams); err != nil {
			return fmt.Errorf("failed to create placeholder meeting: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("DB Creation Error: ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create placeholder meeting")
		return
	}

	// For each attendee, create placeholder attendee row
	for _, attendee := range *body.NewMeeting.Attendees {
		attendeeParams := database.CreatePlaceholderMeetingAttendeeParams{
			//nolint: gosec // id is unsigned 32 bit int
			MeetingID: uint32(placeholderMeeting),
			//nolint: gosec // id is unsigned 32 bit int
			UserID: uint32(attendee),
		}

		err = retry.Do(func() error {
			if _, err = s.DB.CreatePlaceholderMeetingAttendee(ctx, attendeeParams); err != nil {
				return fmt.Errorf("failed to create placeholder meeting attendee: %w", err)
			}
			return nil
		}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	}

	// NOtify user of the request
	SetHeaderAndWriteResponse(w, http.StatusOK, requestID)
}

// (POST /api/reschedule/request/single).
func (s Server) PostAPIRescheduleRequestSingle(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body ReschedulingRequestSingleBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		s.Logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	// Create Rescheduling Request
	var requestID int64
	err = retry.Do(func() error {
		if requestID, err = s.DB.CreateReschedulingRequest(ctx, userID); err != nil {
			return fmt.Errorf("failed to create reschedling requested by user: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))

	// Attach meeting preferences info to request if it exists

	// Link request to old meeting
	var meeting database.Meeting
	// Get data from db to validate meeting id

	meeting, err = s.DB.GetMeetingByMSFTID(ctx, *body.OldMeeting.MsftMeetingID)

	if errors.Is(err, sql.ErrNoRows) {
		newMeetingParams := NewMeetingAndPrefsParams{
			MeetingStartTime: *body.OldMeeting.MeetingStartTime,
			//nolint: gosec // id is unsigned 32 bit int
			OwnerID:       uint32(*body.OldMeeting.MeetingOwner),
			MsftMeetingID: *body.OldMeeting.MsftMeetingID,
		}

		meeting, err = createNewMeetingsAndPrefs(ctx, newMeetingParams, s)
		if err != nil {
			s.Logger.Error("failed to get data from new db.Meeting", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get data from new db.Meeting")
			return
		}
	} else if err != nil {
		s.Logger.Error("failed to get data from db.Meeting", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get data from db.Meeting")
		return
	}

	// Create request to meeting
	requestToMeetingParams := database.CreateRequestToMeetingParams{
		//nolint: gosec // id is unsigned 32 bit int
		RequestID: uint32(requestID),
		MeetingID: meeting.ID,
	}

	err = retry.Do(func() error {
		if _, err = s.DB.CreateRequestToMeeting(ctx, requestToMeetingParams); err != nil {
			return fmt.Errorf("failed to create request to meeting link: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("DB Creation Error: ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create request to meeting link")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, requestID)
}

// (GET /api/reschedule/request).
func (s Server) GetAPIRescheduleRequest(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	// Get all requests for the user
	requests, err := s.DB.GetAllRequestsForUser(ctx, userID)
	if err != nil {
		s.Logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Parse results to response
	response := []RescheduleRequest{}

	for _, req := range requests {
		requestID := int(req.RequestID)
		requestBy := int(req.RequestedBy)

		newMeeting := ReschedulingRequestNewMeeting{}

		if req.MeetingID.Valid {
			newMeeting.EndTime = &req.EndTime.Time
			newMeeting.Location = &req.Location.String

			dur := req.Duration.Time.Format(time.RFC3339Nano)
			newMeeting.MeetingDuration = &dur

			newMeeting.StartTime = &req.StartTime.Time
			newMeeting.Title = &req.Title.String
		}

		meetingID := int(req.ID)

		oldMeeting := ReschedulingRequestOldMeeting{
			MeetingId:     &meetingID,
			MsftMeetingID: &req.MsftMeetingID,
		}

		response = append(response, RescheduleRequest{
			RequestId:   &requestID,
			RequestedAt: &req.CreatedAt,
			RequestedBy: &requestBy,
			Status:      (*string)(&req.Status),
			NewMeeting:  &newMeeting,
			OldMeeting:  &oldMeeting,
		})
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, response)
}

// (GET /api/reschedule/request/{requestID}).
func (s Server) GetAPIRescheduleRequestRequestID(w http.ResponseWriter, r *http.Request, paramRequestID uint32) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	_, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	// Get request for the user
	req, err := s.DB.GetRequestByID(ctx, paramRequestID)
	if err != nil {
		s.Logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Parse results to response

	requestID := int(req.RequestID)
	requestBy := int(req.RequestedBy)

	newMeeting := ReschedulingRequestNewMeeting{}

	if req.MeetingID.Valid {
		newMeeting.EndTime = &req.EndTime.Time
		newMeeting.Location = &req.Location.String

		dur := req.Duration.Time.Format(time.RFC3339Nano)
		newMeeting.MeetingDuration = &dur

		newMeeting.StartTime = &req.StartTime.Time
		newMeeting.Title = &req.Title.String
	}

	meetingID := int(req.ID)

	oldMeeting := ReschedulingRequestOldMeeting{
		MeetingId:     &meetingID,
		MsftMeetingID: &req.MsftMeetingID,
	}

	response := RescheduleRequest{
		RequestId:   &requestID,
		RequestedAt: &req.CreatedAt,
		RequestedBy: &requestBy,
		Status:      (*string)(&req.Status),
		NewMeeting:  &newMeeting,
		OldMeeting:  &oldMeeting,
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, response)
}

// (PATCH /api/reschedule/request/{requestID}/reject).
func (s Server) PatchAPIRescheduleRequestRequestIDReject(w http.ResponseWriter,
	r *http.Request, paramRequestID uint32,
) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	_, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	// Get request for the user
	req, err := s.DB.GetRequestByID(ctx, paramRequestID)
	if err != nil {
		s.Logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Update status of all requests with the same meeting ID
	//nolint: gosec // integer overflow conversion
	_, err = s.DB.UpdateRequestStatusAsRejected(ctx, uint32(req.MeetingID.Int32))
	if err != nil {
		s.Logger.Error("failed to update status of all the requests to declined", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update status of all the requests to declined")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully declined rescheduling request")
}

// (PATCH /api/reschedule/request/{requestID}/accept).
// nolint: funlen // 1 statement too long
func (s Server) PatchAPIRescheduleRequestRequestIDAccept(w http.ResponseWriter, r *http.Request, parRequestID uint32) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body ReschedulingRequestAcceptBodySchema
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

	// Get request for the user
	req, err := s.DB.GetRequestByID(ctx, parRequestID)
	if err != nil {
		s.Logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Update time of calendar event in microsoft
	requestBody := graphmodels.NewEvent()

	timeZone := "UTC"

	start := graphmodels.NewDateTimeTimeZone()
	startTime := body.NewStartTime.Format(time.RFC3339Nano)
	start.SetDateTime(&startTime)
	start.SetTimeZone(&timeZone)
	requestBody.SetStart(start)

	end := graphmodels.NewDateTimeTimeZone()
	endTime := body.NewEndTime.Format(time.RFC3339Nano)
	end.SetDateTime(&endTime)
	end.SetTimeZone(&timeZone)
	requestBody.SetStart(end)

	// Make the microsoft API call
	_, err = graph.Me().Events().ByEventId(req.MsftMeetingID).Patch(ctx, requestBody, nil)
	if err != nil {
		s.Logger.Error("failed to update event in microsoft", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update event in microsoft")
		return
	}

	// Update time in meeting db
	_, err = s.DB.UpdateMeetingStartTime(ctx, database.UpdateMeetingStartTimeParams{
		MeetingStartTime: *body.NewStartTime,
		ID:               req.ID,
	})
	if err != nil {
		s.Logger.Error("failed to update new start time of meeting", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update new start time of meeting")
		return
	}

	// Update status of all requests with the same meeting ID
	//nolint: gosec // integer overflow conversion
	_, err = s.DB.UpdateRequestStatusAsAccepted(ctx, uint32(req.MeetingID.Int32))
	if err != nil {
		s.Logger.Error("failed to update status of all the requests to accepted", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update status of all the requests to accepted")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully accepted rescheduling request")
}
