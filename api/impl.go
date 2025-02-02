package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"golang.org/x/oauth2"
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
	enc.AddInt("id", u.Id)
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
	enc.AddInt("id", t.Id)
	return nil
}

func (tp GetTeamsParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	name := ""
	if tp.Name != nil {
		name = *tp.Name
	}
	enc.AddString("name", name)
	return nil
}

// (GET /healthcheck).
func (s Server) GetHealthcheck(w http.ResponseWriter, _ *http.Request) {
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

// MSFTEntraValues and error.
func getMSFTEntraValues() (MSFTEntraValues, error) {
	var tenantID string
	var clientID string
	var clientSecret string
	var present bool
	// Check if the microsoft entra environment vars are set
	if tenantID, present = os.LookupEnv(TenantIDEnvName); !present {
		return MSFTEntraValues{}, fmt.Errorf("failed to get %s env variable", TenantIDEnvName)
	}

	if clientID, present = os.LookupEnv(ClientIDEnvName); !present {
		return MSFTEntraValues{}, fmt.Errorf("failed to get %s env variable", ClientIDEnvName)
	}

	if clientSecret, present = os.LookupEnv(ClientSecretEnvName); !present {
		return MSFTEntraValues{}, fmt.Errorf("failed to get %s env variable", ClientSecretEnvName)
	}

	return MSFTEntraValues{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
	}, nil
}

type AccessAndRefreshTokens struct {
	AccessToken  string
	RefreshToken string
}

func createAndStoreTokens(userID int, email string, msftRefreshTk string,
	rtr RefreshTokenRepository,
	utmr UserToMSFTRefreshTokenRepository,
) (AccessAndRefreshTokens, error) {
	var accessToken string
	var err error
	if accessToken, err = jwt.CreateNewJWT(userID, email, jwt.AccessTokenJWTSecretEnv); err != nil {
		return AccessAndRefreshTokens{}, fmt.Errorf("failed to create jwt: %w", err)
	}

	// TODO: Put storing and creating in a sql transaction, so if one of those fails then neither are committed
	var refreshToken string
	if refreshToken, err = rtr.CreateRefreshToken(userID, email); err != nil {
		return AccessAndRefreshTokens{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err = utmr.StoreMicrosoftRefreshToken(userID, msftRefreshTk); err != nil {
		return AccessAndRefreshTokens{}, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return AccessAndRefreshTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// the email or it will create a new user and return this user.
func getUserByClaimEmail(claimEmail string, claimName string, ur UserRepository) (User, error) {
	var users Users
	var err error
	users, err = ur.GetUsersByQueryParams(GetUsersParams{
		Email: (*openapi_types.Email)(&claimEmail),
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to get user from claim email: %w", err)
	}

	firstName, lastName := splitName(claimName)

	var u User
	if len(users) == 0 {
		u, err = ur.CreateUser(UserCreate{
			Email:     openapi_types.Email(claimEmail),
			FirstName: firstName,
			LastName:  lastName,
		})
		if err != nil {
			return User{}, fmt.Errorf("failed to create user for claim email: %w", err)
		}
	} else {
		u = users[0]
	}

	return u, nil
}

func (s Server) GetAPIAuthCallback(w http.ResponseWriter, r *http.Request, params GetAPIAuthCallbackParams) {
	msftEntraVals, err := getMSFTEntraValues()
	if err != nil {
		s.Logger.Error("failed to get msft entra values", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "sorry, try again later")
		return
	}

	// Acquire token by authorization code
	discoveryBaseURL := "https://login.microsoftonline.com/organizations/v2.0"
	issuerURL := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", msftEntraVals.TenantID)

	ctx := oidc.InsecureIssuerURLContext(context.Background(), issuerURL)
	provider, err := oidc.NewProvider(ctx, discoveryBaseURL)
	if err != nil {
		s.Logger.Error("failed to create open id provider", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to create open id provider")
		return
	}

	scopes := []string{
		oidc.ScopeOpenID, "profile", "email", "User.ReadWrite",
		"Calendars.ReadBasic", "Calendars.Read", "Calendars.ReadWrite",
		"Calendars.ReadWrite.Shared", "offline_access",
	}

	conf := &oauth2.Config{
		ClientID:     msftEntraVals.ClientID,
		ClientSecret: msftEntraVals.ClientSecret,
		Scopes:       scopes,
		RedirectURL:  "http://localhost:8080/api/auth/callback",
		Endpoint: oauth2.Endpoint{
			TokenURL: provider.Endpoint().TokenURL,
		},
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: HTTPClientTimeOutSecs * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tok, err := conf.Exchange(ctx, params.Code)
	if err != nil {
		s.Logger.Error("failed to get access token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to get access token")
		return
	}
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		s.Logger.Error("failed to get id token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to get id token")
		return
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: msftEntraVals.ClientID})
	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		s.Logger.Error("failed to verify id token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to verify id token")
		return
	}

	// Extract custom claims
	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err = idToken.Claims(&claims); err != nil {
		s.Logger.Error("failed to scan claims", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "failed to scan claims")
		return
	}
	var u User
	if u, err = getUserByClaimEmail(claims.Email, claims.Name, s.UserRepository); err != nil {
		s.Logger.Error("failed to get user for claim email", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "Sorry, try again later")
		return
	}

	var tks AccessAndRefreshTokens
	if tks, err = createAndStoreTokens(u.Id, string(u.Email), tok.RefreshToken,
		s.RefreshTokenRepository, s.UserToMSFTRefreshTokenRepository); err != nil {
		s.Logger.Error("failed to create and store tokens", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "sorry, try again")
		return
	}

	CreateCookies(w, tks.AccessToken, tks.RefreshToken)

	http.Redirect(w, r, "http://localhost:3000/dashboard", http.StatusFound)
}

func (s Server) PostRefresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string
	if refreshToken = r.Header.Get("Refreshtoken"); refreshToken == "" {
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

	var rt jwt.RefreshToken
	if rt, err = s.RefreshTokenRepository.GetRefreshTokenByUserID(userID); err != nil {
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
	var u User
	if u, err = s.UserRepository.GetUserByID(userID); err != nil {
		s.Logger.Error("Failed to refresh token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create new access token
	var accessToken string
	if accessToken, err = jwt.CreateNewJWT(userID, string(u.Email), jwt.AccessTokenJWTSecretEnv); err != nil {
		s.Logger.Error("Failed to refresh token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create new refresh token
	if refreshToken, err = s.RefreshTokenRepository.CreateRefreshToken(userID, string(u.Email)); err != nil {
		s.Logger.Error("Failed to refresh token", zap.Error(err))
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	CreateCookies(w, accessToken, refreshToken)

	SetHeaderAndWriteResponse(w, http.StatusCreated, "Successfully refreshed tokens")
}
