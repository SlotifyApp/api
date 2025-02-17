package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

// Get a group by query params.
// (GET /api/groups)
func (s Server) GetAPIGroups(w http.ResponseWriter, r *http.Request, params GetAPIGroupsParams) {
}

// Get all groups for current user.
// (GET /api/groups/me)
func (s Server) GetAPIGroupsMe(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	// Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	var groups []Group

	userIDStr := strconv.FormatUint(uint64(userID), 10)

	gets, err := graph.Users().ByUserId(userIDStr).MemberOf().Get(ctx, nil)
	if err != nil {
		s.Logger.Error("failed to retrieve memberof of user")
		sendError(w, http.StatusInternalServerError, "Failed to get from microsoft")
		return
	}

	if gets.GetValue() != nil {
		for _, dirs := range gets.GetValue() {
			if grp, ok := dirs.(models.Groupable); ok {
				group, err := GroupableToGroup(grp)
				if err != nil {
					s.Logger.Error("failed to convert groupable to group")
					sendError(w, http.StatusInternalServerError, "Failed to convert groupable to groups")
				}
				groups = append(groups, group)
			}
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, groups)

}

// Get a group by id.
// (GET /api/groups/{groupID})
func (s Server) GetAPIGroupsGroupID(w http.ResponseWriter, r *http.Request, groupID uint32) {

}

// Get all members of a group.
// (GET /api/groups/{groupID}/users)
func (s Server) GetAPIGroupsGroupIDUsers(w http.ResponseWriter, r *http.Request, groupID uint32) {

}
