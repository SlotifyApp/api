package api

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request, params GetAPICalendarMeParams) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	//Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	//Create msgraph client
	graph, err := CreateMSFTGraphClient(ctx, s.Logger, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	//Make call to API route and parse events
	calendarEvents, err := makeCalendarMeAPICall(graph, params.Start, params.End)
	if err != nil {
		s.Logger.Error("failed to make calendar me msgraph api call", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to make calendar me msgraph api call")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}
