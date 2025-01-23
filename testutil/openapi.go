// Package testutil provides testing utilities.
package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/stretchr/testify/require"
	validator "openapi.tanna.dev/go/validator/openapi3"
)

// OpenAPIValidateTestHelper is a test helper to
// ensure the request and response matches the OpenAPI spec.
func OpenAPIValidateTestHelper(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
	t.Run("it matches OpenAPI", func(t *testing.T) {
		doc, err := api.GetSwagger()
		require.NoError(t, err, "GetSwagger doesn't return error")

		_ = validator.NewValidator(doc).ForTest(t, rr, req)
	})
}
