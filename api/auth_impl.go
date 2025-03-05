package api

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/jwt"
	"go.uber.org/zap"
)

// (POST /api/refresh).
func (s Server) PostAPIRefresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*database.DatabaseTimeout)
	defer cancel()

	refreshToken, ok := r.Context().Value(RefreshTokenCtxKey{}).(string)
	reqUUID, _ := ReadReqUUID(r)
	uuidStr := zap.String("request ID: ", reqUUID)
	if !ok {
		s.Logger.Error("failed to parse refresh token from context value into string, ", uuidStr)
		sendError(w, http.StatusUnauthorized, "failed to parse refresh token from context value into string")
		return
	}

	claims, err := jwt.ParseJWT(refreshToken, jwt.RefreshTokenJWTSecretEnv)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "refresh token was invalid")
		return
	}

	userID := claims.UserID

	var rt database.RefreshToken
	if rt, err = s.DB.GetRefreshTokenByUserID(ctx, userID); err != nil {
		sendError(w, http.StatusUnauthorized, "failed to refresh token")
		return
	}

	// check if the actual user's refresh token matches the request's refresh token
	if rt.Token != refreshToken || rt.Revoked {
		s.Logger.Error("Failed to match provided token or verify token OR token was revoked", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "failed to refresh token")
		return
	}

	// Generate new access token and new refresh token
	var uq database.User
	if uq, err = s.DB.GetUserByID(ctx, userID); err != nil {
		s.Logger.Error("Failed to refresh token, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var tks jwt.AccessAndRefreshTokens
	tks, err = jwt.CreateAccessAndRefreshTokens(ctx, s.Logger, &s.DB.Queries, userID, uq.Email)
	if err != nil {
		s.Logger.Error("failed to create access and refresh tokens, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to create access and refresh tokens")
		return
	}

	CreateCookies(w, tks.AccessToken, tks.RefreshToken)

	SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully refreshed tokens")
}

// (GET /api/auth/callback).
func (s Server) GetAPIAuthCallback(w http.ResponseWriter, r *http.Request, params GetAPIAuthCallbackParams) {
	msftTokenRes, err := msftAuthoriseByCode(r.Context(), s.MSALClient, params.Code)
	reqUUID, _ := ReadReqUUID(r)
	uuidStr := zap.String("request ID: ", reqUUID)
	if err != nil {
		s.Logger.Error("failed to get microsoft tokens, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Sorry, try again later. Failed to get Microsoft tokens.")
		return
	}

	tx, err := s.DB.DB.Begin()
	if err != nil {
		s.Logger.Error("failed to start db transaction, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusInternalServerError, "callback route: failed to start db transaction")
		return
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			s.Logger.Error("failed to rollback db transaction, ", uuidStr, zap.Error(err))
		}
	}()

	qtx := s.DB.WithTx(tx)
	var u database.User
	if u, err = getOrInsertUserByClaimEmail(r.Context(), qtx, msftTokenRes); err != nil {
		s.Logger.Error("failed to get user for claim email from msft access token, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusBadRequest, "failed to parse msft access token")
		return
	}

	var tks jwt.AccessAndRefreshTokens
	if tks, err = jwt.CreateAccessAndRefreshTokens(r.Context(), s.Logger, qtx, u.ID, u.Email); err != nil {
		s.Logger.Error("failed to create and store tokens, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to create slotify access and refresh token")
		return
	}

	if err = tx.Commit(); err != nil {
		s.Logger.Error("failed to commit db transaction, ", uuidStr, zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to commit db transaction")
		return
	}

	CreateCookies(w, tks.AccessToken, tks.RefreshToken)

	frontendURL, present := os.LookupEnv("FRONTEND_URL")
	if !present {
		s.Logger.Error("failed to get FRONTEND_URL value, request ID: ", uuidStr)
		sendError(w, http.StatusInternalServerError, "Sorry, failed to get required env var")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/dashboard", frontendURL), http.StatusFound)
}
