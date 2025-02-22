package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// (POST /api/scheduling/free).
func (s Server) PostAPISchedulingFree(w http.ResponseWriter, r *http.Request) {
	// Get userid from access token
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	var body SchedulingSlotsSuccessResponseBody
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

	// Get custom graph request header and config
	graphConfigAndBody, err := CreateSchedulingGraphReqBody()
	if err != nil {
		SetHeaderAndWriteResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to make scheduling call: %s", err.Error()))
		return
	}
	findMeetingTimes, err := graph.Me().FindMeetingTimes().Post(ctx, graphConfigAndBody.reqBody, graphConfigAndBody.config)
	if err != nil {
		SetHeaderAndWriteResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to make scheduling call: %s", err.Error()))
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, findMeetingTimes)
}
