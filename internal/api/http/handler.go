package http

import (
	"github.com/polyxia-org/agent/internal/api"
	"github.com/polyxia-org/agent/internal/runtime"
	"net/http"
)

func (s *server) invokeHandler(w http.ResponseWriter, r *http.Request) {
	invokeRequest, err := parseRequestBody[api.InvokeRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
		// TODO: message
	}

	// Validate the invocation request DTO
	if err := invokeRequest.Validate(); err != nil {

		s.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
		// TODO: message
	}

	// Parse query arguments into runtime.FnVars
	vars := runtime.FnVars{}
	for k := range r.URL.Query() {
		vars[k] = r.URL.Query().Get(k)
	}

	result, err := s.ex.Invoke(invokeRequest, vars)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: message
		return
	}

	s.JSONResponse(w, result)
}
