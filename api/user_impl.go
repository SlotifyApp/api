package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SlotifyApp/slotify-backend/database"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

// (GET /users) Get a user by query params.
func (s Server) GetAPIUsers(w http.ResponseWriter, r *http.Request, params GetAPIUsersParams) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	users, err := s.DB.ListUsers(ctx, database.ListUsersParams{
		Email:     params.Email,
		FirstName: params.FirstName,
		LastName:  params.LastName,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("user api: failed to get users: context cancelled")
			sendError(w, http.StatusInternalServerError, "user api: failed to get users")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("user api: failed to get users: query timed out")
			sendError(w, http.StatusInternalServerError, "user api: failed to get users")
			return
		default:
			s.Logger.Error("user api: failed to get users")
			sendError(w, http.StatusInternalServerError, "user api: failed to get users")
			return
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}

// (POST /users) Create a new user.
func (s Server) PostAPIUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	var userBody PostAPIUsersJSONRequestBody
	var err error
	defer func() {
		if err = r.Body.Close(); err != nil {
			s.Logger.Warn("could not close request body", zap.Error(err))
		}
	}()
	if err = json.NewDecoder(r.Body).Decode(&userBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Object("body", userBody), zap.Error(err))
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
			s.Logger.Error("user api: context cancelled", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("user api: query timed out", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "DB query timed out")

		case database.IsDuplicateEntrySQLError(err):
			s.Logger.Error("user api: user already exists", zap.Object("req_body", userBody), zap.Error(err))
			sendError(w, http.StatusBadRequest, fmt.Sprintf("user with email %s already exists", userBody.Email))
		default:
			s.Logger.Error("user api failed to create user", zap.Object("req_body", userBody), zap.Error(err))
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
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	rowsDeleted, err := s.DB.DeleteUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("user api: context cancelled", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("user api: query timed out", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: query timed out")
			return
		default:
			s.Logger.Error("user api failed to delete user", zap.Uint32("userID", userID), zap.Error(err))
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
		s.Logger.Error(errMsg, zap.Uint32("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "user deleted successfully")
}

// (GET /users/{userID}) Get a user by id.
func (s Server) GetAPIUsersUserID(w http.ResponseWriter, r *http.Request, userID uint32) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	dbUser, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Error("user api: context cancelled", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")

		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Error("user api: query timed out", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api: query timed out")

		case errors.Is(err, sql.ErrNoRows):
			errMsg := fmt.Sprintf("user api: user with id(%d) doesn't exist", userID)
			s.Logger.Error(errMsg, zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusForbidden, errMsg)

		default:
			errMsg := fmt.Sprintf("user api: failed to get user with id(%d)", userID)
			s.Logger.Error(errMsg, zap.Uint32("userID", userID), zap.Error(err))
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
	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	s.GetAPIUsersUserID(w, r, userID)
}

// (POST /users/logout).
func (s Server) PostAPIUsersMeLogout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), database.DatabaseTimeout)
	defer cancel()

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		RemoveCookies(w)
		// still logout successfully, dont return error on logout
		SetHeaderAndWriteResponse(w, http.StatusOK, "Logging out")
		return
	}
	// Remove refresh token from db
	rowsDeleted, err := s.DB.DeleteRefreshTokenByUserID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			s.Logger.Errorf("user api: logout query context cancelled", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError, "user api: context cancelled")
			return
		case errors.Is(err, context.DeadlineExceeded):
			s.Logger.Errorf("user api: logout query timed out", zap.Uint32("userID", userID))
			sendError(w, http.StatusInternalServerError, "user api: query timed out")
			return
		default:
			s.Logger.Errorf("user api: failed to logout user", zap.Uint32("userID", userID), zap.Error(err))
			sendError(w, http.StatusInternalServerError, "user api failed to delete user")
			return
		}
	}

	if rowsDeleted != 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsDeleted,
			ExpectedRows: []int64{1},
		}
		s.Logger.Errorf("user api failed to logout user", zap.Error(err))
	}

	RemoveCookies(w)
	SetHeaderAndWriteResponse(w, http.StatusOK, "Logging out")
}
