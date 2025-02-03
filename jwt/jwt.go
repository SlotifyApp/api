package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	goJWT "github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenJWTSecretEnv = "ACCESS_TOKEN_JWT_SECRET"
	//nolint: gosec //This doesn't leak anything, it's just the env var name
	RefreshTokenJWTSecretEnv = "REFRESH_TOKEN_JWT_SECRET"
	OneWeek                  = 7 * 24 * time.Hour
)

var (
	ErrNoAuthHeader      = errors.New("header Authorization is missing")
	ErrInvalidAuthHeader = errors.New("header Authorization is malformed")
)

// CustomClaims is a struct for Slotify JWT claims.
type CustomClaims struct {
	UserID uint32 `json:"user_id"`
	goJWT.RegisteredClaims
}

// CreateNewJWT returns a signed JWT.
func CreateNewJWT(userID uint32, email string, keyEnv string) (string, error) {
	m := map[string]time.Duration{
		AccessTokenJWTSecretEnv:  time.Hour,
		RefreshTokenJWTSecretEnv: OneWeek,
	}
	var expiryDur time.Duration
	var ok bool
	if expiryDur, ok = m[keyEnv]; !ok {
		return "", errors.New("keyEnv did not have expiry set")
	}
	key, present := os.LookupEnv(keyEnv)
	if !present {
		return "", fmt.Errorf("failed to create jwt token: %s env var missing", AccessTokenJWTSecretEnv)
	}
	t := goJWT.NewWithClaims(goJWT.SigningMethodHS512,
		CustomClaims{
			RegisteredClaims: goJWT.RegisteredClaims{
				Issuer:    "slotify",
				Subject:   email,
				ExpiresAt: goJWT.NewNumericDate(time.Now().Add(expiryDur)),
				IssuedAt:  goJWT.NewNumericDate(time.Now()),
			},
			UserID: userID,
		},
	)
	signedToken, err := t.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create jwt token: %w", err)
	}
	return signedToken, nil
}

// ParseJWT verifies and parses whether the token is valid.
func ParseJWT(tk string, keyEnv string) (CustomClaims, error) {
	allowedKeyEnvs := map[string]struct{}{
		AccessTokenJWTSecretEnv:  {},
		RefreshTokenJWTSecretEnv: {},
	}

	if _, ok := allowedKeyEnvs[keyEnv]; !ok {
		return CustomClaims{}, errors.New("key env not part of allowed key sources")
	}

	key, present := os.LookupEnv(keyEnv)
	if !present {
		return CustomClaims{}, fmt.Errorf("failed to parse jwt: %s env var missing", AccessTokenJWTSecretEnv)
	}

	token, err := goJWT.ParseWithClaims(tk, &CustomClaims{}, func(token *goJWT.Token) (interface{}, error) {
		if _, ok := token.Method.(*goJWT.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to parse token: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		return CustomClaims{}, fmt.Errorf("failed to parse jwt: %w", err)
	}

	var claims *CustomClaims
	var ok bool

	if claims, ok = token.Claims.(*CustomClaims); ok && token.Valid {
		return *claims, nil
	}

	return CustomClaims{}, fmt.Errorf("failed to parse jwt, token valid: %t", token.Valid)
}

// GetJWTFromRequest extracts a JWT string from an Authorization: Bearer <jwt> header.
func GetJWTFromRequest(req *http.Request) (string, error) {
	authHdr := req.Header.Get("Authorization")
	// Check for the Authorization header.
	if authHdr == "" {
		return "", ErrNoAuthHeader
	}
	// We expect a header value of the form "Bearer <token>", with 1 space after
	// Bearer, per spec.
	prefix := "Bearer "
	if !strings.HasPrefix(authHdr, prefix) {
		return "", ErrInvalidAuthHeader
	}
	return strings.TrimPrefix(authHdr, prefix), nil
}

// GetUserIDFromReq gets the user id from the request Authorization header.
func GetUserIDFromReq(r *http.Request) (uint32, error) {
	var err error
	var accessToken string
	if accessToken, err = GetJWTFromRequest(r); err != nil {
		return 0, fmt.Errorf("failed to getuserid from req: %w", err)
	}
	var claims CustomClaims
	if claims, err = ParseJWT(accessToken, AccessTokenJWTSecretEnv); err != nil {
		return 0, fmt.Errorf("failed to getuserid from req: %w", err)
	}

	return claims.UserID, nil
}

// CreateNewRefreshToken will generate a new refresh token and store it in the database.
func CreateNewRefreshToken(ctx context.Context, db *database.Database, userID uint32, email string) (string, error) {
	// TODO: Put storing and creating in a sql transaction, so if one of those fails then neither are committed
	refreshToken, err := CreateNewJWT(userID, email, RefreshTokenJWTSecretEnv)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	dbParams := database.CreateRefreshTokenParams{
		UserID: userID,
		Token:  refreshToken,
	}

	rowsAffected, err := db.CreateRefreshToken(ctx, dbParams)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return "", fmt.Errorf("context cancelled during create refresh token: %w", err)
		case errors.Is(err, context.DeadlineExceeded):
			return "", fmt.Errorf("create refresh token timed out: %w", err)
		default:
			return "", fmt.Errorf("failed to create refresh token: %w", err)
		}
	}

	if rowsAffected != 1 {
		return "", database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{1},
		}
	}

	return refreshToken, nil
}
