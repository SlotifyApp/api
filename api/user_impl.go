package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

const (
	// SearchUserLim is the limit for number of users to return from the users search.
	SearchUserLim = 10
)

// (GET /users) Get a user by query params.
func (s Server) GetAPIUsers(w http.ResponseWriter, r *http.Request, params GetAPIUsersParams) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	users := make([]database.User, 0)

	// searches by name first if both name and email params are given a value
	switch {
	// search users by name
	case params.Name != nil:
		usersByName, err := s.DB.SearchUsersByName(ctx,
			//nolint: gosec // page is unsigned 32 bit int
			database.SearchUsersByNameParams{
				Name:   *params.Name,
				Limit:  SearchUserLim,
				LastID: uint32(params.PageToken),
			})
		if err != nil {
			logger.Error("failed to search user by name", zap.Error(err),
				zap.String("name", *params.Name))
			sendError(w, http.StatusInternalServerError, "failed to search users by name")
			return
		}
		for _, row := range usersByName {
			users = append(users, database.User{
				ID:        row.ID,
				Email:     row.Email,
				FirstName: row.FirstName,
				LastName:  row.LastName,
			})
		}
	// search users by email
	case params.Email != nil:
		usersByEmail, err := s.DB.SearchUsersByEmail(ctx,
			//nolint: gosec // page is unsigned 32 bit int
			database.SearchUsersByEmailParams{
				Email:  string(*params.Email),
				Limit:  SearchUserLim,
				LastID: uint32(params.PageToken),
			})
		if err != nil {
			logger.Error("failed to search user by email", zap.Error(err),
				zap.String("email", string(*params.Email)))
			sendError(w, http.StatusInternalServerError, "failed to search users by email")
			return
		}
		for _, row := range usersByEmail {
			users = append(users, database.User{
				ID:        row.ID,
				Email:     row.Email,
				FirstName: row.FirstName,
				LastName:  row.LastName,
			})
		}
	// Default, just list all users
	default:
		usersAll, err := s.DB.ListUsers(ctx,
			//nolint: gosec // page is unsigned 32 bit int
			database.ListUsersParams{
				Limit:  SearchUserLim,
				LastID: uint32(params.PageToken),
			})
		if err != nil {
			logger.Error("failed to list all users", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "failed to list all users")
			return
		}
		for _, row := range usersAll {
			users = append(users, database.User{
				ID:        row.ID,
				Email:     row.Email,
				FirstName: row.FirstName,
				LastName:  row.LastName,
			})
		}
	}
	var nextPageToken int
	if len(users) == SearchUserLim {
		// id of the last user gotten here
		nextPageToken = int(users[len(users)-1].ID)
	} else {
		// -1 if end of database
		nextPageToken = -1
	}

	response := struct {
		Users         []database.User
		NextPageToken int
	}{
		Users:         users,
		NextPageToken: nextPageToken,
	}
	SetHeaderAndWriteResponse(w, http.StatusOK, response)
}

// (POST /users) Create a new user.
func (s Server) PostAPIUsers(w http.ResponseWriter, r *http.Request) {
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)

	defer cancel()

	var userBody PostAPIUsersJSONRequestBody
	var err error
	defer func() {
		if err = r.Body.Close(); err != nil {
			logger.Warn("could not close request body", zap.Error(err))
		}
	}()
	if err = json.NewDecoder(r.Body).Decode(&userBody); err != nil {
		logger.Error(ErrUnmarshalBody, zap.Object("body", userBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	var userID int64
	userID, err = s.DB.CreateUser(ctx, database.CreateUserParams{
		Email:     string(userBody.Email),
		FirstName: userBody.FirstName,
		LastName:  userBody.LastName,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logger.Error("user api: context cancelled", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")
		case errors.Is(err, context.DeadlineExceeded):
			logger.Error("user api: query timed out", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "DB query timed out")

		case database.IsDuplicateEntrySQLError(err):
			logger.Error("user api: user already exists", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusBadRequest, fmt.Sprintf("user with email %s already exists", userBody.Email))
		default:
			logger.Error("user api failed to create user", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api failed to create user")
		}
		return
	}

	u := User{
		//nolint: gosec // id is unsigned 32 bit int
		Id:        uint32(userID),
		FirstName: userBody.FirstName,
		LastName:  userBody.LastName,
		Email:     userBody.Email,
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, u)
}

// (DELETE /users/{userID}) Delete a user by id.
func (s Server) DeleteAPIUsersUserID(w http.ResponseWriter, r *http.Request, userID uint32) {
	loggerInUserID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("logged_in_user_id", loggerInUserID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)

	defer cancel()

	rowsDeleted, err := s.DB.DeleteUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logger.Error("user api: context cancelled", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			logger.Error("user api: query timed out", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: query timed out")
			return
		default:
			logger.Error("user api failed to delete user", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api failed to delete user")
			return
		}
	}

	if rowsDeleted != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsDeleted,
			ExpectedRows: []int64{1},
		}
		errMsg := fmt.Sprintf("user api: user with id(%d) doesn't exist", userID)
		logger.Error(errMsg, zap.Uint32("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "user deleted successfully")
}

// (GET /users/{userID}) Get a user by id.
func (s Server) GetAPIUsersUserID(w http.ResponseWriter, r *http.Request, userID uint32) {
	loggedInUserID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("logged_in_user_id", loggedInUserID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	dbUser, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logger.Error("user api: context cancelled", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")

		case errors.Is(err, context.DeadlineExceeded):
			logger.Error("user api: query timed out", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: query timed out")

		case errors.Is(err, sql.ErrNoRows):
			errMsg := fmt.Sprintf("user api: user with id(%d) doesn't exist", userID)
			logger.Error(errMsg, zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusNotFound, errMsg)

		default:
			errMsg := fmt.Sprintf("user api: failed to get user with id(%d)", userID)
			logger.Error(errMsg, zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusBadRequest, errMsg)
		}
		return
	}

	u := User{
		Id:        dbUser.ID,
		Email:     openapi_types.Email(dbUser.Email),
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, u)
}

// (GET /users/me).
func (s Server) GetAPIUsersMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)

	s.GetAPIUsersUserID(w, r, userID)
}

// (POST /users/logout).
func (s Server) PostAPIUsersMeLogout(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	// Remove refresh token from db
	rowsDeleted, err := s.DB.DeleteRefreshTokenByUserID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logger.Errorf("user api: logout query context cancelled", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			logger.Errorf("user api: logout query timed out", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError, "user api: query timed out")
			return
		default:
			logger.Errorf("user api: failed to logout user", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api failed to delete user")
			return
		}
	}

	if rowsDeleted != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsDeleted,
			ExpectedRows: []int64{1},
		}
		logger.Errorf("user api failed to logout user", zap.Error(err))
	}

	RemoveCookies(w)
	SetHeaderAndWriteResponse(w, http.StatusOK, "Logging out")
}

// (GET /api/msft-users/).
func (s Server) GetAPIMSFTUsers(w http.ResponseWriter, r *http.Request) {
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

	groupable, err := graph.Users().Get(context.Background(), nil)
	if err != nil {
		logger.Errorf("failed to get users from microsoft: %v", err)
		sendError(w, http.StatusNotFound, fmt.Sprintf("Failed to find users: %v", err))
		return
	}

	var users []MSFTUser

	if groupable.GetValue() != nil {
		for _, usr := range groupable.GetValue() {
			var user MSFTUser
			user, err = UserableToMSFTUser(usr)
			if err != nil {
				logger.Errorf("failed to convert userable to user: %v", err)
				sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to convert userable to user: %v", err))
				return
			}
			users = append(users, user)
		}
	}
	// no error 204 since we got an array

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}

// if param is empty then calls GetAPIMSFTUsers.
func (s Server) GetAPIMSFTUsersSearch(w http.ResponseWriter, r *http.Request, params GetAPIMSFTUsersSearchParams) {
	userID, _ := r.Context().Value(UserIDCtxKey{}).(uint32)
	reqID, _ := r.Context().Value(RequestIDCtxKey{}).(string)

	logger := s.Logger.With(zap.String("request_id", reqID), zap.Uint32("user_id", userID))

	if params.Search == nil {
		s.GetAPIMSFTUsers(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)
	if err != nil {
		logger.Error("failed to create msgraph client", zap.Error(err))
		sendError(w, http.StatusBadGateway, "Failed to connect to microsoft graph API")
		return
	}

	var requestParameters *graphusers.ItemPeopleRequestBuilderGetQueryParameters

	requestSearch := fmt.Sprintf("\"%s\"", *params.Search)
	requestParameters = &graphusers.ItemPeopleRequestBuilderGetQueryParameters{
		Search: &requestSearch,
	}

	configuration := &graphusers.ItemPeopleRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	groupable, err := graph.Me().People().Get(context.Background(), configuration)
	if err != nil {
		logger.Error("failed to get persons from microsoft")
		sendError(w, http.StatusNotFound, "Failed to find group")
		return
	}

	var users []MSFTUser

	if groupable.GetValue() != nil {
		for _, usr := range groupable.GetValue() {
			var user MSFTUser
			user, err = PersonableToMSFTUser(usr)
			if err != nil {
				logger.Error("failed to convert userable to user")
				sendError(w, http.StatusInternalServerError, "Failed to convert userable to user")
				return
			}
			users = append(users, user)
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}
