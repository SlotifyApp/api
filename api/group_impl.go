package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SlotifyApp/slotify-backend/database"
	graphgroups "github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

// Get a group by query params.
// (GET /api/groups)
func (s Server) GetAPIGroups(w http.ResponseWriter, r *http.Request, params GetAPIGroupsParams) {

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

	filter := fmt.Sprintf("displayName eq '%s'", params.Name)
	configuration := &graphgroups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: &graphgroups.GroupsRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	}

	gets, err := graph.Groups().Get(ctx, configuration)
	if err != nil {
		s.Logger.Error("failed to get group")
		sendError(w, http.StatusInternalServerError, "Failed to connect to get group")
		return
	}

	if gets.GetValue() == nil || len(gets.GetValue()) == 0 {
		s.Logger.Error("no group found")
		sendError(w, http.StatusNotFound, "Failed to find a group with name")
		return
	}

	groupable, ok := gets.GetValue()[0].(models.Groupable)
	if !ok {
		s.Logger.Error("failed to cast to groupable")
		sendError(w, http.StatusInternalServerError, "Failed to cast to groupable")
		return
	}

	group, err := GroupableToGroup(groupable)
	if err != nil {
		s.Logger.Error("error converting groupable")
		sendError(w, http.StatusInternalServerError, "Failed to convert groupable")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, group)
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
					return
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

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Get(ctx, nil)
	if err != nil {
		s.Logger.Error("failed to get group from microsoft")
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var group Group

	if groupable != nil && groupable.GetId() != nil && groupable.GetDisplayName() != nil {
		group, err = GroupableToGroup(groupable)
		if err != nil {
			s.Logger.Error("error converting groupable")
			sendError(w, http.StatusInternalServerError, "Failed to convert groupable")
			return
		}
	} else {
		s.Logger.Error("got invalid group from microsoft")
		sendError(w, http.StatusInternalServerError, "Invalid group")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, group)
}

// Get all members of a group.
// (GET /api/groups/{groupID}/users)
func (s Server) GetAPIGroupsGroupIDUsers(w http.ResponseWriter, r *http.Request, groupID uint32) {

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

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Members().Get(ctx, nil)
	if err != nil {
		s.Logger.Error("failed to get group from microsoft")
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var users []User

	if groupable.GetValue() != nil {
		for _, dirs := range groupable.GetValue() {
			if usr, ok := dirs.(models.Userable); ok {
				user, err := UserableToUser(usr)
				if err != nil {
					s.Logger.Error("failed to convert userable to user")
					sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
					return
				}
				users = append(users, user)
			}
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}
