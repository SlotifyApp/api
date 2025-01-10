package api

import (
	"encoding/json"
	"net/http"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check.
var _ ServerInterface = (*Server)(nil)

// (GET /healthcheck).
func (s Server) GetHealthcheck(w http.ResponseWriter, _ *http.Request) {
	resp := "Healthcheck Successful!"

	s.Logger.Info("Healthcheck successful through logger!")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
