package api

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// (GET /api/rooms/all).
func (s Server) GetAPIRoomsAll(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	graphRoom, err := graph.Places().GraphRoom().Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get response from msgraph", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get response from msgraph for list of rooms")
		return
	}

	rooms := graphRoom.GetValue()
	if rooms == nil {
		logger.Error("graph room GetValue was nil", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to get response from msgraph for list of rooms")
		return
	}

	parsedRooms := make([]Room, 0)
	for _, r := range rooms {
		email := r.GetEmailAddress()
		if email == nil {
			logger.Error("email for room was nil")
			sendError(w, http.StatusBadRequest,
				"Failed to get room's email address")
			return
		}

		displayName := "placeholder name"
		if r.GetDisplayName() != nil {
			displayName = *r.GetDisplayName()
		}
		parsedRooms = append(parsedRooms, Room{
			Email: openapi_types.Email(*email),
			Name:  displayName,
		})
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, parsedRooms)
}
