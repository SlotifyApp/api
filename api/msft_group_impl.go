package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SlotifyApp/slotify-backend/database"
	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
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

	var outs []string
	for _, st := range gets.GetValue() {
		outs = append(outs, st)
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, outs)
}

// (GET /api/msft-groups/{groupID}).
func (s Server) GetAPIMSFTGroupsGroupID(w http.ResponseWriter, r *http.Request, groupID uint32) {
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

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Members().Get(ctx, nil)
	if err != nil {
		logger.Error("failed to get group from microsoft", zap.Error(err))
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var users []MSFTUser

	pageIterator, err := graphcore.NewPageIterator[models.DirectoryObjectable](
		groupable,
		graph.GetAdapter(),
		models.CreateDirectoryObjectCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		logger.Error("failed to initiate page iterator", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to initiate page iterator")
		return
	}

	err = pageIterator.Iterate(ctx, func(d models.DirectoryObjectable) bool {
		if usr, ok := d.(models.Userable); ok {
			user, convErr := UserableToMSFTUser(usr)
			if convErr != nil {
				logger.Error("failed to convert userable to user")
				sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
				return false
			}
			users = append(users, user)
		}
		return true
	})
	if err != nil {
		logger.Error("failed to iterate group member pages", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Failed to iterate group member pages")
		return
	}

	// if groupable.GetValue() != nil {
	// 	for _, dirs := range groupable.GetValue() {
	// 		if usr, ok := dirs.(models.Userable); ok {
	// 			var user MSFTUser
	// 			user, err = UserableToMSFTUser(usr)
	// 			if err != nil {
	// 				logger.Error("failed to convert userable to user")
	// 				sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
	// 				return
	// 			}
	// 			users = append(users, user)
	// 		}
	// 	}
	// }

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}
