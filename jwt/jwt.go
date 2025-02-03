package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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
	UserID int `json:"user_id"`
	goJWT.RegisteredClaims
}

// RefreshToken is a struct for Slotify refresh tokens.
type RefreshToken struct {
	ID      int
	UserID  int
	Token   string
	Revoked bool
}

// CreateNewJWT returns a signed JWT.
func CreateNewJWT(userID int, email string, keyEnv string) (string, error) {
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
func GetUserIDFromReq(r *http.Request) (int, error) {
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
