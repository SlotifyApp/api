package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
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

// (GET /healthcheck).
func (s Server) GetAPIHealthcheck(w http.ResponseWriter, r *http.Request) {
	resp := "Healthcheck Successful!"
	w.WriteHeader(http.StatusOK)
	reqUUID := ReadReqUUID(r)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.Logger.Error("Failed to encode JSON, request ID: "+reqUUID+", ", zap.String("body", resp))
		sendError(w, http.StatusInternalServerError, "Failed to encode JSON")
		return
	}
}
