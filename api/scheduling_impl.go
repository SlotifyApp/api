package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// (POST /api/scheduling/free).
func (s Server) PostAPISchedulingFree(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*3)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqUUID := ReadReqUUID(r)
	if !ok {
		s.Logger.Error("failed to get userid from request context, request ID: " + reqUUID)
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body SchedulingSlotsBodySchema
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		// TODO: Add zap log for body
		s.Logger.Error(ErrUnmarshalBody.Error()+", request ID: "+reqUUID+", ", zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client, request ID: "+reqUUID+", ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	respBody, err := makeFindMeetingTimesAPICall(ctx, graph, body)
	if err != nil {
		s.Logger.Error("failed to make msgraph api call to findMeetings, request ID: "+reqUUID+", ", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to process/send microsoft graph API request for findMeeting")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, respBody)
}
