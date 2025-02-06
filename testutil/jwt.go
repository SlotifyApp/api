package testutil

import (
	"testing"

	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/stretchr/testify/require"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func CreateJWT(t *testing.T, userID uint32, email openapi_types.Email) string {
	jwt, err := jwt.GenerateJWT(userID, string(email), jwt.AccessTokenJWTSecretEnv)

	require.NoError(t, err, "failed to create jwt token")

	return jwt
}
