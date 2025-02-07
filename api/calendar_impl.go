package api

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// (GET /calendar/me).
func (s Server) GetAPICalendarMe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	userID, ok := r.Context().Value(UserCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	at, err := getMSFTAccessToken(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to get microsoft access token", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := createMSFTGraphClient(at)

	if err != nil || graph == nil {
		s.Logger.Error("failed to create msft graph client", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "msft graph could not be created")
		return
	}

	events, err := graph.Me().Calendar().Events().Get(ctx, nil)
	if err != nil {
		s.Logger.Error("failed to call graph client route", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to call graph client route")
		return
	}
	calendarEvents := []CalendarEvent{}
	for _, e := range events.GetValue() {
		ce := CalendarEvent{
			StartTime: e.GetStart().GetDateTime(),
			EndTime:   e.GetEnd().GetDateTime(),
			Subject:   e.GetSubject(),
		}
		calendarEvents = append(calendarEvents, ce)
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, calendarEvents)
}
