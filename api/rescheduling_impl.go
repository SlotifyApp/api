package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
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

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	resBody, err := processReschedulingRequest(ctx, graph, body)

	if err != nil {
		s.Logger.Error("failed to create request", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to create request")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, resBody)
}
