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

// (GET /api/msft-groups).
func (s Server) GetAPIMSFTGroups(w http.ResponseWriter, r *http.Request, params GetAPIMSFTGroupsParams) {
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With("request_id", reqID)

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	// Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	filter := fmt.Sprintf("displayName eq '%s'", *params.Name)
	configuration := &graphgroups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: &graphgroups.GroupsRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	}

	gets, err := graph.Groups().Get(ctx, configuration)
	if err != nil {
		logger.Error("failed to get group")
		sendError(w, http.StatusInternalServerError, "Failed to connect to get group")
		return
	}

	if gets.GetValue() == nil || len(gets.GetValue()) == 0 {
		logger.Error("no group found")
		sendError(w, http.StatusNotFound, "Failed to find a group with name")
		return
	}

	// uses the first group
	group, err := GroupableToMSFTGroup(gets.GetValue()[0])
	if err != nil {
		logger.Error("error converting groupable")
		sendError(w, http.StatusInternalServerError, "Failed to convert groupable")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, group)
}

// (GET  /api/msft-groups/me).
func (s Server) GetAPIMSFTGroupsMe(w http.ResponseWriter, r *http.Request) {
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With("request_id", reqID)

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	// Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	userIDStr := strconv.FormatUint(uint64(userID), 10)

	gets, err := graph.Users().ByUserId(userIDStr).MemberOf().Get(ctx, nil)
	if err != nil {
		logger.Error("failed to retrieve memberof of user")
		sendError(w, http.StatusInternalServerError, "Failed to get from microsoft")
		return
	}

	groups, err := GetsToMSFTGroups(gets)
	if err != nil {
		logger.Error("failed to retrive groups", zap.Error(err))
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed retrieving groups: %v", err))
		return
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, groups)
}

// (GET /api/msft-groups/{groupID}).
func (s Server) GetAPIMSFTGroupsGroupID(w http.ResponseWriter, r *http.Request, groupID uint32) {
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With("request_id", reqID)

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	// Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get group from microsoft")
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var group MSFTGroup

	if groupable != nil && groupable.GetId() != nil && groupable.GetDisplayName() != nil {
		group, err = GroupableToMSFTGroup(groupable)
		if err != nil {
			logger.Error("error converting groupable")
			sendError(w, http.StatusInternalServerError, "Failed to convert groupable")
			return
		}
	} else {
		logger.Error("got invalid group from microsoft")
		sendError(w, http.StatusInternalServerError, "Invalid group")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, group)
}

// (GET /api/msft-groups/{groupID}/users).
func (s Server) GetAPIMSFTGroupsGroupIDUsers(w http.ResponseWriter, r *http.Request, groupID uint32) {
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)
	logger := s.Logger.With("request_id", reqID)

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	// Get userID from request
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Members().Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get group from microsoft")
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var users []MSFTGroupUser

	if groupable.GetValue() != nil {
		for _, dirs := range groupable.GetValue() {
			if usr, ok := dirs.(models.Userable); ok {
				var user MSFTGroupUser
				user, err = UserableToMSFTGroupUser(usr)
				if err != nil {
					logger.Error("failed to convert userable to user")
					sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
					return
				}
				users = append(users, user)
			}
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}
