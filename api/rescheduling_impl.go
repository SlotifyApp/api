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

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	var body ReschedulingCheckBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Get old meeting data from microsoft
	msftMeeting, err := graph.Me().Events().ByEventId(body.OldMeeting.MsftMeetingID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get meeting data from microsoft", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get meeting data from microsoft")
		return
	}

	// Get data from db
	var meeting database.Meeting
	meeting, err = s.DB.GetMeetingByMSFTID(ctx, body.OldMeeting.MsftMeetingID)

	var meetingFound bool

	if errors.Is(err, sql.ErrNoRows) {
		// Meeting Not Found
		meetingFound = false
	} else if err != nil {
		logger.Error("failed to search meeting table in db", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process db request")
		return
	}

	var meetingPref database.Meetingpreferences
	if meetingFound {
		// Get meeting preferences if data exists
		meetingPref, err = s.DB.GetMeetingPreferences(ctx, meeting.MeetingPrefID)
		if err != nil {
			logger.Error("failed to search meeting table in db", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to process db request")
			return
		}
	} else {
		// Create temp meeting preferences if data doesn't exist

		dayTime := time.Hour * hoursInAWeek // 1 week : 24 * 7

		var loc *time.Location
		loc, err = time.LoadLocation(*msftMeeting.GetStart().GetTimeZone())
		if err != nil {
			logger.Error("failed to parse start time zone", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to parse start time zone")
			return
		}

		var newStartTime time.Time
		newStartTime, err = time.ParseInLocation(time.RFC3339Nano, *msftMeeting.GetStart().GetDateTime()+"Z", loc)
		if err != nil {
			logger.Error("failed to parse start time", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to parse start time")
			return
		}

		meetingPref = database.Meetingpreferences{
			StartDateRange: time.Now(),
			EndDateRange:   newStartTime.Add(dayTime), // Give a week extra from the start of the meeting
		}
	}

	respBody, err := performReschedulingCheckProcess(ctx, graph, body, msftMeeting, meetingPref)
	if err != nil {
		logger.Error("failed to make msgraph api call to findMeetings", zap.Error(err))
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

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	var body ReschedulingRequestBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Create Rescheduling Request
	var requestID int64
	err = retry.Do(func() error {
		if requestID, err = s.DB.CreateReschedulingRequest(ctx, database.CreateReschedulingRequestParams{
			RequestedBy: userID,
			CreatedAt:   time.Now(),
		}); err != nil {
			return fmt.Errorf("failed to create reschedling requested by user: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))

	// Attach meeting preferences info to request if it exists

	// Link request to old meeting
	var meeting database.Meeting
	// Get data from db to validate meeting id

	meeting, err = s.DB.GetMeetingByMSFTID(ctx, body.OldMeeting.MsftMeetingID)

	if errors.Is(err, sql.ErrNoRows) {
		// Meeting info not in db, so create new meeting info
		meeting, err = processNewMeetingInfo(ctx, graph, s, body.OldMeeting.MsftMeetingID, *logger)
		if err != nil {
			logger.Error("DB Creation Error: ", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to create New Meeting Info")
			return
		}
	} else if err != nil {
		logger.Error("failed to get data from db.Meeting", zap.Error(err))
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
		logger.Error("DB Creation Error: ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create request to meeting link")
		return
	}

	// Create Placeholder meeting info
	parsedTime, err := time.Parse(time.RFC3339Nano, body.NewMeeting.MeetingDuration)
	if err != nil {
		logger.Error("failed to parse time", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to parse time")
		return
	}

	placeholderParams := database.CreatePlaceholderMeetingParams{
		//nolint: gosec // id is unsigned 32 bit int
		RequestID: uint32(requestID),
		Title:     body.NewMeeting.Title,
		StartTime: body.NewMeeting.StartTime,
		EndTime:   body.NewMeeting.EndTime,
		Location:  body.NewMeeting.Location,

		Duration:       parsedTime,
		StartDateRange: body.NewMeeting.StartRangeTime,
		EndDateRange:   body.NewMeeting.EndRangeTime,
	}

	var placeholderMeeting int64
	err = retry.Do(func() error {
		if placeholderMeeting, err = s.DB.CreatePlaceholderMeeting(ctx, placeholderParams); err != nil {
			return fmt.Errorf("failed to create placeholder meeting: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		logger.Error("DB Creation Error: ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create placeholder meeting")
		return
	}

	// For each attendee, create placeholder attendee row
	for _, attendee := range body.NewMeeting.Attendees {
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

	// Notify user of the request
	notifParam := database.CreateNotificationParams{
		Message: "Reschedule request for meeting for a new meeting",
		Created: time.Now(),
	}

	// Get Owner ID
	ownerObj, err := s.DB.GetUserByEmail(ctx, meeting.OwnerEmail)
	if err != nil {
		logger.Error("Failed to get owner user obj: ", zap.Error(err))
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{ownerObj.ID}, notifParam)
	if err != nil {
		logger.Error("Failed to send notification for reschedule request: ", zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, requestID)
}

// (POST /api/reschedule/request/single).
// nolint: funlen // 1 line too long
func (s Server) PostAPIRescheduleRequestSingle(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	var body ReschedulingRequestSingleBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Create Rescheduling Request
	var requestID int64
	err = retry.Do(func() error {
		if requestID, err = s.DB.CreateReschedulingRequest(ctx, database.CreateReschedulingRequestParams{
			RequestedBy: userID,
			CreatedAt:   time.Now(),
		}); err != nil {
			return fmt.Errorf("failed to create reschedling requested by user: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		logger.Error("failed to make reschedule request", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to make rescheduling request")
		return
	}

	// Link request to old meeting
	var meeting database.Meeting

	// Get data from db to validate meeting id
	meeting, err = s.DB.GetMeetingByMSFTID(ctx, body.MsftMeetingID)

	if errors.Is(err, sql.ErrNoRows) {
		meeting, err = processNewMeetingInfo(ctx, graph, s, body.MsftMeetingID, *logger)
		if err != nil {
			logger.Error("failed to make new meeting info", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to make new meeting info")
			return
		}
	} else if err != nil {
		logger.Error("failed to get data from db.Meeting", zap.Error(err))
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
		logger.Error("DB Creation Error: ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create request to meeting link")
		return
	}

	// Notify user of the request
	notifParam := database.CreateNotificationParams{
		Message: "Reschedule request for meeting",
		Created: time.Now(),
	}

	// Get Owner ID
	ownerObj, err := s.DB.GetUserByEmail(ctx, meeting.OwnerEmail)
	if err != nil {
		logger.Error("Failed to get owner user obj: ", zap.Error(err))
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{ownerObj.ID}, notifParam)
	if err != nil {
		logger.Error("Failed to send notification for reschedule request: ", zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, requestID)
}

// (GET /api/reschedule/requests/me).
// nolint: funlen // by 5 lines
func (s Server) GetAPIRescheduleRequestsMe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	// Get all requests for the user
	userObj, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("failed to get user data from db", zap.Error(err))
		sendError(w, http.StatusBadGateway, "failed to get user data from db")
		return
	}

	requests, err := s.DB.GetAllRequestsForOwner(ctx, userObj.Email)
	if err != nil {
		logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Parse results to response
	response := []RescheduleRequest{}
	// nolint: dupl // Duplicate code
	for _, req := range requests {
		newMeeting := ReschedulingRequestNewMeeting{}

		if req.MeetingID.Valid {
			newMeeting.EndTime = req.EndTime.Time
			newMeeting.Location = req.Location.String

			dur := req.Duration.Time.Format(time.RFC3339Nano)
			newMeeting.MeetingDuration = dur

			newMeeting.StartTime = req.StartTime.Time
			newMeeting.Title = req.Title.String

			var attendeeIDs []uint32
			// nolint: gosec // int to uint
			attendeeIDs, err = s.DB.GetPlaceholderMeetingAttendeesByMeetingID(ctx, uint32(req.MeetingID.Int32))
			if err != nil {
				logger.Error("failed to get requests", zap.Error(err))
				sendError(w, http.StatusBadGateway, "Failed to get requests")
				return
			}

			newMeeting.Attendees = attendeeIDs
		}

		response = append(response, RescheduleRequest{
			RequestId:   req.RequestID,
			RequestedAt: req.CreatedAt,
			RequestedBy: req.RequestedBy,
			Status:      string(req.Status),
			NewMeeting:  &newMeeting,
			OldMeeting: ReschedulingRequestOldMeeting{
				MeetingId:        req.ID,
				MsftMeetingID:    req.MsftMeetingID,
				MeetingStartTime: req.MeetingStartTime,
				TimeRangeStart:   req.StartDateRange,
				TimeRangeEnd:     req.EndDateRange,
			},
		})
	}

	respondedReqs, err := s.DB.GetAllRequestsResponsesForUserID(ctx, userID)
	if err != nil {
		logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Parse results to response
	respondedReqsResponses := []RescheduleRequest{}
	// nolint: dupl // Duplicate code
	for _, req := range respondedReqs {
		newMeeting := ReschedulingRequestNewMeeting{}

		if req.MeetingID.Valid {
			newMeeting.EndTime = req.EndTime.Time
			newMeeting.Location = req.Location.String

			dur := req.Duration.Time.Format(time.RFC3339Nano)
			newMeeting.MeetingDuration = dur

			newMeeting.StartTime = req.StartTime.Time
			newMeeting.Title = req.Title.String

			var attendeeIDs []uint32
			// nolint: gosec // int to uint
			attendeeIDs, err = s.DB.GetPlaceholderMeetingAttendeesByMeetingID(ctx, uint32(req.MeetingID.Int32))
			if err != nil {
				logger.Error("failed to get requests", zap.Error(err))
				sendError(w, http.StatusBadGateway, "Failed to get requests")
				return
			}

			newMeeting.Attendees = attendeeIDs
		}

		respondedReqsResponses = append(respondedReqsResponses, RescheduleRequest{
			RequestId:   req.RequestID,
			RequestedAt: req.CreatedAt,
			RequestedBy: req.RequestedBy,
			Status:      string(req.Status),
			NewMeeting:  &newMeeting,
			OldMeeting: ReschedulingRequestOldMeeting{
				MeetingId:        req.ID,
				MsftMeetingID:    req.MsftMeetingID,
				MeetingStartTime: req.MeetingStartTime,
				TimeRangeStart:   req.StartDateRange,
				TimeRangeEnd:     req.EndDateRange,
			},
		})
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, RescheduleRequests{
		Pending:   response,
		Responses: respondedReqsResponses,
	})
}

// (GET /api/reschedule/request/{requestID}).
func (s Server) GetAPIRescheduleRequestRequestID(w http.ResponseWriter, r *http.Request, paramRequestID uint32) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	// Get request for the user
	req, err := s.DB.GetRequestByID(ctx, paramRequestID)
	if err != nil {
		logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Parse results to response
	newMeeting := ReschedulingRequestNewMeeting{}

	if req.MeetingID.Valid {
		newMeeting.EndTime = req.EndTime.Time
		newMeeting.Location = req.Location.String

		dur := req.Duration.Time.Format(time.RFC3339Nano)
		newMeeting.MeetingDuration = dur

		newMeeting.StartTime = req.StartTime.Time
		newMeeting.Title = req.Title.String

		var attendeeIDs []uint32
		// nolint: gosec // int to uint
		attendeeIDs, err = s.DB.GetPlaceholderMeetingAttendeesByMeetingID(ctx, uint32(req.MeetingID.Int32))
		if err != nil {
			logger.Error("failed to get requests", zap.Error(err))
			sendError(w, http.StatusBadGateway, "Failed to get requests")
			return
		}

		newMeeting.Attendees = attendeeIDs
	}

	response := RescheduleRequest{
		RequestId:   req.RequestID,
		RequestedAt: req.CreatedAt,
		RequestedBy: req.RequestedBy,
		Status:      string(req.Status),
		NewMeeting:  &newMeeting,
		OldMeeting: ReschedulingRequestOldMeeting{
			MeetingId:        req.ID,
			MsftMeetingID:    req.MsftMeetingID,
			MeetingStartTime: req.MeetingStartTime,
			TimeRangeStart:   req.StartDateRange,
			TimeRangeEnd:     req.EndDateRange,
		},
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

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	// Get request for the user
	req, err := s.DB.GetMeetingIDFromRequestID(ctx, paramRequestID)
	if err != nil {
		logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Update status of all requests with the same meeting ID

	_, err = s.DB.UpdateRequestStatusAsRejected(ctx, req.MeetingID)
	if err != nil {
		logger.Error("failed to update status of all the requests to declined", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update status of all the requests to declined")
		return
	}

	// Notify user of the request
	notifParam := database.CreateNotificationParams{
		Message: "Reschedule request rejected",
		Created: time.Now(),
	}

	// Get request owner
	requester, err := s.DB.GetOnlyRequestByID(ctx, req.RequestID)
	if err != nil {
		logger.Error("Failed to get requester user: ", zap.Error(err))
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{requester.RequestedBy}, notifParam)
	if err != nil {
		logger.Error("Failed to send notification to requester: ", zap.Error(err))
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, notifParam)
	if err != nil {
		logger.Error("Failed to send notification to requester: ", zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully declined rescheduling request")
}

// (PATCH /api/reschedule/request/{requestID}/accept).
// nolint: funlen // 1 statement too long
func (s Server) PatchAPIRescheduleRequestRequestIDAccept(w http.ResponseWriter, r *http.Request, parRequestID uint32) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	var body ReschedulingRequestAcceptBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	// Get request for the user
	req, err := s.DB.GetRequestByID(ctx, parRequestID)
	if err != nil {
		logger.Error("failed to get requests", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get requests")
		return
	}

	// Update time of calendar event in microsoft
	requestBody := graphmodels.NewEvent()

	timeZone := "GMT Standard Time"

	start := graphmodels.NewDateTimeTimeZone()
	startTime := body.NewStartTime.Format(time.RFC3339Nano)
	start.SetDateTime(&startTime)
	start.SetTimeZone(&timeZone)
	requestBody.SetStart(start)

	end := graphmodels.NewDateTimeTimeZone()
	endTime := body.NewEndTime.Format(time.RFC3339Nano)
	end.SetDateTime(&endTime)
	end.SetTimeZone(&timeZone)
	requestBody.SetEnd(end)

	// Make the microsoft API call
	_, err = graph.Me().Events().ByEventId(req.MsftMeetingID).Patch(ctx, requestBody, nil)
	if err != nil {
		logger.Error("failed to update event in microsoft", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update event in microsoft")
		return
	}

	// Update time in meeting db
	_, err = s.DB.UpdateMeetingStartTime(ctx, database.UpdateMeetingStartTimeParams{
		MeetingStartTime: body.NewStartTime,
		ID:               req.ID,
	})
	if err != nil {
		logger.Error("failed to update new start time of meeting", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update new start time of meeting")
		return
	}

	// Update status of all requests with the same meeting ID

	rows, err := s.DB.UpdateRequestStatusAsAccepted(ctx, req.ID)
	if err != nil {
		logger.Error("failed to update status of all the requests to accepted", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to update status of all the requests to accepted")
		return
	}

	// At least 1 row has to be updated
	if rows == 0 {
		logger.Error("failed to update a request invite as accepted",
			zap.Error(database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}))
		sendError(w, http.StatusBadGateway, "Failed to update invite message")
		return
	}

	notifparam := database.CreateNotificationParams{
		Message: "You have successfully rescheduled",
		Created: time.Now(),
	}

	// Notify Owner
	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, []uint32{userID}, notifparam)
	if err != nil {
		logger.Error("failed to send accepted request notification to owner", zap.Error(err))
	}

	// Notify all attendees

	meetingData, err := graph.Me().Events().ByEventId(req.MsftMeetingID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get meeting data from microsoft", zap.Error(err))
	}

	var attendees []Attendee
	if attendees, err = parseMSFTAttendees(meetingData); err != nil {
		logger.Error("failed to parse msft attendees", zap.Error(err))
		sendError(w, http.StatusBadRequest, "failed to parse the msft attendees received")
		return
	}
	attendeeIDs := make([]uint32, 0)
	for _, a := range attendees {
		var u database.User
		// don't send error, we only need the id for notifications, just log
		if u, err = s.DB.GetUserByEmail(ctx, string(a.Email)); err != nil {
			logger.Error("failed to get user by email in PatchAPIReschedule for sending notifications",
				zap.Error(err))
		}
		attendeeIDs = append(attendeeIDs, u.ID)
	}

	newNotifparam := database.CreateNotificationParams{
		Message: "Meeting x has been rescheduled",
		Created: time.Now(),
	}

	err = s.NotificationService.SendNotification(ctx, s.Logger, s.DB, attendeeIDs, newNotifparam)
	if err != nil {
		logger.Error("failed to send accepted request notification to all attendees", zap.Error(err))
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully accepted rescheduling request")
}

// (POST /api/reschedule/request/{requesteID}/complete).
func (s Server) PostAPIRescheduleRequestRequestIDComplete(w http.ResponseWriter, r *http.Request, parRequestID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	s.PostAPICalendarMe(w, r)

	if r.Response.StatusCode == http.StatusOK {
		// Delete request as event has been created for it
		err := s.DB.DeleteRequest(ctx, parRequestID)
		if err != nil {
			logger.Error("failed to delete request after creating new event", zap.Error(err))
		}
	}
}

// (GET /api/reschedule/request/{requesteID}/close).
func (s Server) GetAPIRescheduleRequestRequestIDClose(w http.ResponseWriter, r *http.Request, parRequestID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("userID", userID))

	// Delete request as event has been created for it
	err := s.DB.DeleteRequest(ctx, parRequestID)
	if err != nil {
		logger.Error("failed to delete request after creating new event", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to delete request")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "Successfully deleted request")
}
