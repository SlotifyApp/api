package api

import (
	"net/http"
)

// (POST /api/scheduling/free).
func (s Server) PostAPISchedulingFree(w http.ResponseWriter, _ *http.Request) {
	SetHeaderAndWriteResponse(w, http.StatusOK, "Hello")
}
