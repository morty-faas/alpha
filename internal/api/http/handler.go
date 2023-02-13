package http

import (
	"context"
	"github.com/google/uuid"
	"github.com/polyxia-org/agent/internal/api"
	"github.com/polyxia-org/agent/internal/runtime"
	"net/http"
)

func (s *server) invokeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), "iid", uuid.New().String())

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
	params := runtime.FnParams{}
	for k := range r.URL.Query() {
		params[k] = r.URL.Query().Get(k)
	}

	result, err := s.ex.Invoke(ctx, &runtime.FnInvocation{
		CodeURL: invokeRequest.Code,
		Runtime: invokeRequest.Runtime,
		Params:  params,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: message
		return
	}

	s.JSONResponse(w, result)
}
