package http

import (
	"github.com/polyxia-org/agent/internal/api"
	"net/http"
)

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Currently we don't have any health probes implemented.
	// We can return by default an HTTP 200 response.
	// But in a future version, we will need to define how our agent
	// can be marked as ready by the system to be able to trigger functions.
	s.JSONResponse(w, &api.HealthResponse{
		Status: "UP",
	})
}
