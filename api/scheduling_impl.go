package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// (POST /api/scheduling/free).
func (s Server) PostAPISchedulingFree(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	reqID, userID, err := GetCtxValues(r)
	if err != nil {
		if errors.Is(err, ErrRequestIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusInternalServerError, "Try again later.")
		} else if errors.Is(err, ErrUserIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusUnauthorized, "Try again later.")
		}
		return
	}

	var body SchedulingSlotsBodySchema
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		s.Logger.Errorf(ErrUnmarshalBody.Error(), zap.String("request id", reqID), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("%s: %s: failed to create msgraph client, ", reqID, zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	respBody, err := makeFindMeetingTimesAPICall(ctx, graph, body)
	if err != nil {
		s.Logger.Error("failed to make msgraph api call to findMeetings, ", reqID, zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process/send microsoft graph API request for findMeeting")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, respBody)
}
