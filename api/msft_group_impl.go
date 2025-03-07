package api

import (
	"context"
	"errors"
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
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	reqID, userID, err := GetCtxValues(r)
	if err != nil {
		if errors.Is(err, ErrRequestIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusInternalServerError, "Try again later.")
		} else if errors.Is(err, ErrUserIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusUnauthorized, "Try again later.")
		}
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Errorf("%s: failed to create msgraph client, ", reqID, zap.Error(err))
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
		s.Logger.Errorf("%s: failed to get group, ", reqID)
		sendError(w, http.StatusInternalServerError, "Failed to connect to get group")
		return
	}

	if gets.GetValue() == nil || len(gets.GetValue()) == 0 {
		s.Logger.Errorf("%s: no group found, ", reqID)
		sendError(w, http.StatusNotFound, "Failed to find a group with name")
		return
	}

	// uses the first group
	group, err := GroupableToMSFTGroup(gets.GetValue()[0])
	if err != nil {
		s.Logger.Errorf("%s: error converting groupable, ", reqID)
		sendError(w, http.StatusInternalServerError, "Failed to convert groupable")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, group)
}

// (GET  /api/msft-groups/me).
func (s Server) GetAPIMSFTGroupsMe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	reqID, userID, err := GetCtxValues(r)
	if err != nil {
		if errors.Is(err, ErrRequestIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusInternalServerError, "Try again later.")
		} else if errors.Is(err, ErrUserIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusUnauthorized, "Try again later.")
		}
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Errorf("%s: failed to create msgraph client, ", reqID, zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	userIDStr := strconv.FormatUint(uint64(userID), 10)

	gets, err := graph.Users().ByUserId(userIDStr).MemberOf().Get(ctx, nil)
	if err != nil {
		s.Logger.Errorf("%s: failed to retrieve memberof of user, ", reqID)
		sendError(w, http.StatusInternalServerError, "Failed to get from microsoft")
		return
	}

	groups, err := GetsToMSFTGroups(gets)
	if err != nil {
		s.Logger.Errorf("%s: failed to retrive , ", reqID, zap.Error(err))
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed retrieving groups: %v", err))
		return
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, groups)
}

// (GET /api/msft-groups/{groupID}).
func (s Server) GetAPIMSFTGroupsGroupID(w http.ResponseWriter, r *http.Request, groupID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	reqID, userID, err := GetCtxValues(r)
	if err != nil {
		if errors.Is(err, ErrRequestIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusInternalServerError, "Try again later.")
		} else if errors.Is(err, ErrUserIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusUnauthorized, "Try again later.")
		}
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Errorf("%s: failed to create msgraph client, ", reqID, zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Get(ctx, nil)
	if err != nil {
		s.Logger.Errorf("%s: failed to get group from microsoft, ", reqID)
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var group MSFTGroup

	if groupable != nil && groupable.GetId() != nil && groupable.GetDisplayName() != nil {
		group, err = GroupableToMSFTGroup(groupable)
		if err != nil {
			s.Logger.Errorf("%s: error converting groupable, ", reqID)
			sendError(w, http.StatusInternalServerError, "Failed to convert groupable")
			return
		}
	} else {
		s.Logger.Errorf("%s: got invalid group from microsoft, ", reqID)
		sendError(w, http.StatusInternalServerError, "Invalid group")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, group)
}

// (GET /api/msft-groups/{groupID}/users).
func (s Server) GetAPIMSFTGroupsGroupIDUsers(w http.ResponseWriter, r *http.Request, groupID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	reqID, userID, err := GetCtxValues(r)
	if err != nil {
		if errors.Is(err, ErrRequestIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusInternalServerError, "Try again later.")
		} else if errors.Is(err, ErrUserIDNotFound) {
			s.Logger.Error(err)
			sendError(w, http.StatusUnauthorized, "Try again later.")
		}
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Errorf("%s: failed to create msgraph client, ", reqID, zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	groupIDStr := strconv.FormatUint(uint64(groupID), 10)

	groupable, err := graph.Groups().ByGroupId(groupIDStr).Members().Get(ctx, nil)
	if err != nil {
		s.Logger.Errorf("%s: failed to get group from microsoft, ", reqID)
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
					s.Logger.Errorf("%s: failed to convert userable to user, ", reqID)
					sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
					return
				}
				users = append(users, user)
			}
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}
