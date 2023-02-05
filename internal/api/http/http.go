package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/polyxia-org/agent/internal/executor"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type (
	server struct {
		logger *log.Logger
		ex     *executor.Executor
	}
)

func NewServer(ex *executor.Executor) *server {
	logger := log.New()
	return &server{
		logger,
		ex,
	}
}

func (s *server) Serve() error {
	r := chi.NewRouter()
	r.Post("/", s.invokeHandler)

	serv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:3000"),
		Handler: r,
	}

	s.logger.Info("HTTP API Server is listening on 0.0.0.0:3000")
	if err := serv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func parseRequestBody[T interface{}](r *http.Request) (*T, error) {
	defer r.Body.Close()
	result := new(T)
	if err := json.NewDecoder(r.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *server) JSONResponse(w http.ResponseWriter, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("JSON Marshal failed : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(prettyJson(body))
}

func prettyJson(b []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, b, "", " ")
	return out.Bytes()
}
