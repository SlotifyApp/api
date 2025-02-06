package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/avast/retry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HTTPClientTimeOutSecs = 2
	PrepareStmtFail       = "failed to prepare sql stmt"
	QueryDBFail           = "failed to query database"
	TenantIDEnvName       = "MICROSOFT_TENANT_ID"
	ClientIDEnvName       = "MICROSOFT_CLIENT_ID"
	ClientSecretEnvName   = "MICROSOFT_CLIENT_SECRET"
)

var ErrUnmarshalBody = errors.New("failed to unmarshal request body correctly")

// ensure that we've conformed to the `ServerInterface` with a compile-time check.
var _ ServerInterface = (*Server)(nil)

// sendError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(message)
}

// Set JSON content-type header and send response.
func SetHeaderAndWriteResponse(w http.ResponseWriter, code int, encode any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(encode); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to encode JSON")
	}
}

type (
	Users []User
	Teams []Team
)

// MarshalLogArray implements the zapcore.ArrayMarshaler interface.
func (u Users) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, user := range u {
		if err := arr.AppendObject(user); err != nil {
			return err
		}
	}
	return nil
}

// MarshalLogObject implements zapcore.ObjectMarshaler.

func (u UserCreate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("email", string(u.Email))
	enc.AddString("firstName", u.FirstName)
	enc.AddString("lastName", u.LastName)
	return nil
}

func (u User) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	userCreate := UserCreate{
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
	if err := userCreate.MarshalLogObject(enc); err != nil {
		return fmt.Errorf("failed to marshal User obj: %v", err.Error())
	}
	enc.AddUint32("id", u.Id)
	return nil
}

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (t TeamCreate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", t.Name)
	return nil
}

// MarshalLogArray implements the zapcore.ArrayMarshaler interface.
func (t Teams) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, team := range t {
		if err := arr.AppendObject(team); err != nil {
			return err
		}
	}
	return nil
}

func (t Team) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	teamCreate := TeamCreate{
		Name: t.Name,
	}
	if err := teamCreate.MarshalLogObject(enc); err != nil {
		return fmt.Errorf("failed to marshal Team obj: %v", err.Error())
	}
	enc.AddUint32("id", t.Id)
	return nil
}

func (tp GetAPITeamsParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	name := ""
	if tp.Name != nil {
		name = *tp.Name
	}
	enc.AddString("name", name)
	return nil
}

// (GET /healthcheck).
func (s Server) GetAPIHealthcheck(w http.ResponseWriter, _ *http.Request) {
	resp := "Healthcheck Successful!"
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.Logger.Error("Failed to encode JSON", zap.String("body", resp))
		sendError(w, http.StatusInternalServerError, "Failed to encode JSON")
		return
	}
}

// splitName splits a name into first name and last name.
func splitName(name string) (string, string) {
	names := strings.Fields(name)
	var firstName string
	var lastName string

	if len(names) > 0 {
		firstName = names[0]
	}

	if len(names) > 1 {
		// Join the rest of the fields together to form the
		// last name
		lastName = strings.Join(names[1:], " ")
	}
	return firstName, lastName
}

type MSFTEntraValues struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

type AccessAndRefreshTokens struct {
	AccessToken  string
	RefreshToken string
}

// createAndStoreTokens will generate an access and refresh token and store the refresh token.
func createAndStoreTokens(qtx database.Queries, userID uint32, email string) (AccessAndRefreshTokens, error) {
	var accessToken string
	var err error
	if accessToken, err = jwt.GenerateJWT(userID, email, jwt.AccessTokenJWTSecretEnv); err != nil {
		return AccessAndRefreshTokens{}, fmt.Errorf("failed to create jwt: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.DatabaseTimeout)
	defer cancel()

	var refreshToken string
	refreshToken, err = jwt.GenerateAndStoreRefreshToken(ctx, &qtx, userID, email)
	if err != nil {
		return AccessAndRefreshTokens{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return AccessAndRefreshTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// getUserByClaimEmail will get user by the claim email or if first time log in, it will create a new user.
func getUserByClaimEmail(qtx *database.Queries, msftTokenRes MSFTTokenResult) (database.User, error) {
	email := msftTokenRes.Email
	// Double the timeout due to more db operations
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
	defer cancel()

	count, err := qtx.CountUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, fmt.Errorf("failed to get user count by claim email: %w", err)
	}

	// User doesn't exist, first time signing so create a new user
	if count == 0 {
		dbParams := database.CreateUserParams{
			Email:     email,
			FirstName: msftTokenRes.FirstName,
			LastName:  msftTokenRes.LastName,
		}
		_, err = qtx.CreateUser(ctx, dbParams)
		if err != nil {
			return database.User{}, fmt.Errorf("failed to create user for claim email: %w", err)
		}
	}

	var u database.User

	u, err = qtx.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, fmt.Errorf("failed to get user with claim email: %w", err)
	}

	// Update user's home account id so it can be used when asking for a MSFT access token
	dbParams := database.UpdateUserHomeAccountIDParams{
		ID:                u.ID,
		MsftHomeAccountID: sql.NullString{String: msftTokenRes.HomeAccountID, Valid: true},
	}
	var rowsAffected int64
	rowsAffected, err = qtx.UpdateUserHomeAccountID(ctx, dbParams)
	if err != nil {
		return database.User{}, fmt.Errorf("failed to update user home account id: %w", err)
	}

	// UpdateUserHomeAccountID will either update 1 or 0 rows.
	if rowsAffected > 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{0, 1},
		}
		return database.User{}, fmt.Errorf("failed to update home account id: %w", err)
	}

	return u, nil
}

func (s Server) GetAPIAuthCallback(w http.ResponseWriter, r *http.Request, params GetAPIAuthCallbackParams) {
	msftTokenRes, err := msftAuthoriseByCode(context.Background(), s.MSALClient, params.Code)
	if err != nil {
		s.Logger.Error("failed to get microsoft tokens", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Sorry, try again later. Failed to get Microsoft tokens.")
		return
	}

	tx, err := s.DB.DB.Begin()
	if err != nil {
		s.Logger.Error("failed to start db transaction", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "callback route: failed to start db transaction")
		return
	}

	defer func() {
		// TODO: Check condition
		if err = tx.Rollback(); err != nil {
			s.Logger.Error("failed to rollback db transaction", zap.Error(err))
		}
	}()
	qtx := s.DB.WithTx(tx)

	var tks AccessAndRefreshTokens
	err = retry.Do(func() error {
		var u database.User
		if u, err = getUserByClaimEmail(qtx, msftTokenRes); err != nil {
			s.Logger.Error("failed to get user for claim email", zap.Error(err))
			return err
		}

		if tks, err = createAndStoreTokens(*qtx, u.ID, u.Email); err != nil {
			s.Logger.Error("failed to create and store tokens", zap.Error(err))
			return err
		}

		if err = tx.Commit(); err != nil {
			s.Logger.Error("failed to commit db transaction", zap.Error(err))
			return err
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("All retries to get user by claim email and store tokens failed", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Sorry, try again later.")
		return
	}

	CreateCookies(w, tks.AccessToken, tks.RefreshToken)

	http.Redirect(w, r, "http://localhost:3000/dashboard", http.StatusFound)
}

func (s Server) PostAPIRefresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string
	if refreshToken = r.Header.Get(refreshTokenHeader); refreshToken == "" {
		s.Logger.Error("refresh token was empty")
		sendError(w, http.StatusUnauthorized, "refresh token was empty")
		return
	}
	claims, err := jwt.ParseJWT(refreshToken, jwt.RefreshTokenJWTSecretEnv)
	if err != nil {
		s.Logger.Errorf("failed to verify refreshToken", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "refresh token was invalid")
		return
	}

	userID := claims.UserID

	ctx, cancel := context.WithTimeout(context.Background(), 2*database.DatabaseTimeout)
	defer cancel()

	var rt database.RefreshToken
	if rt, err = s.DB.GetRefreshTokenByUserID(ctx, userID); err != nil {
		s.Logger.Error("failed to get refresh token for user", zap.Error(err))
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
		s.Logger.Error("Failed to refresh token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create new access token
	var accessToken string
	if accessToken, err = jwt.GenerateJWT(userID, uq.Email, jwt.AccessTokenJWTSecretEnv); err != nil {
		s.Logger.Error("Failed to refresh token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var newRefreshToken string
	tx, err := s.DB.DB.Begin()
	if err != nil {
		s.Logger.Error("Failed to start db transaction", zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			s.Logger.Error("failed to rollback db transaction")
		}
	}()
	qtx := s.DB.WithTx(tx)

	err = retry.Do(func() error {
		// Create new refresh token
		if newRefreshToken, err = jwt.GenerateAndStoreRefreshToken(ctx, qtx, userID, uq.Email); err != nil {
			s.Logger.Error("Failed to refresh token", zap.Error(err))
			return err
		}

		if err = tx.Commit(); err != nil {
			s.Logger.Error("Failed to commit db transaction", zap.Error(err))
			return err
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		s.Logger.Error("all retries to generate and store token failed", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Sorry, try again.")
	}

	CreateCookies(w, accessToken, newRefreshToken)

	SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully refreshed tokens")
}
