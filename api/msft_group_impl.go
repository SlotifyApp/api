package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	graphgroups "github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
	"go.uber.org/zap"
)

// (GET /api/msft-groups).
func (s Server) GetAPIMSFTGroups(w http.ResponseWriter, r *http.Request, params GetAPIMSFTGroupsParams) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

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
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	requestBody := graphusers.NewItemGetMemberGroupsPostRequestBody()
	securityEnabledOnly := true
	requestBody.SetSecurityEnabledOnly(&securityEnabledOnly)

	gets, err := graph.Me().GetMemberGroups().PostAsGetMemberGroupsPostResponse(
		ctx,
		requestBody,
		nil,
	)
	if err != nil {
		logger.Error("failed to fetch my groups from Microsoft", zap.Error(err))
		sendError(w, http.StatusNotFound, "Failed to fetch my groups from Microsoft")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, gets.GetValue())
}

// (GET /api/msft-groups/{groupID}).
func (s Server) GetAPIMSFTGroupsGroupID(w http.ResponseWriter, r *http.Request, groupID string) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	groupable, err := graph.Groups().ByGroupId(groupID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get group from microsoft")
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var group MSFTGroup

	if groupable != nil && groupable.GetId() != nil && groupable.GetDisplayName() != nil {
		group, err = GroupableToMSFTGroup(groupable)
		if err != nil {
			logger.Error("error converting groupable", zap.Error(err))
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
func (s Server) GetAPIMSFTGroupsGroupIDUsers(w http.ResponseWriter, r *http.Request, groupID string,
	params GetAPIMSFTGroupsGroupIDUsersParams,
) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	var users []MSFTUser

	requestParameters := &graphgroups.ItemMembersRequestBuilderGetQueryParameters{
		Top: &params.Limit,
	}
	configuration := &graphgroups.ItemMembersRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	var gets models.DirectoryObjectCollectionResponseable
	// initial, when there is no nextLink url
	if params.NextLink == nil {
		gets, err = graph.Groups().ByGroupId(groupID).Members().Get(ctx, configuration)
	} else {
		gets, err = graph.Groups().ByGroupId(groupID).Members().WithUrl(*params.NextLink).Get(ctx, configuration)
	}

	if err != nil {
		logger.Error("failed to get group from microsoft", zap.Error(err))
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	if gets.GetValue() != nil {
		for _, dirs := range gets.GetValue() {
			if usr, ok := dirs.(models.Userable); ok {
				var user MSFTUser
				user, err = UserableToMSFTUser(usr)
				if err != nil {
					logger.Error("failed to convert userable to user")
					sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
					return
				}
				users = append(users, user)
			}
		}
	}

	nextLink := gets.GetOdataNextLink()

	resp := struct {
		Users    []MSFTUser `json:"users"`
		NextLink *string    `json:"nextLink,omitempty"`
	}{
		Users:    users,
		NextLink: nextLink,
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, resp)
}
