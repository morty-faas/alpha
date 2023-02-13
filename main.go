package main

import (
	httpapi "github.com/polyxia-org/agent/internal/api/http"
	"github.com/polyxia-org/agent/internal/runtime"
	log "github.com/sirupsen/logrus"
)

func main() {
	exec, err := runtime.New()
	if err != nil {
		log.Errorf("failed to initialize runtime : %v", exec)
	}

	httpapi.
		NewServer(exec).
		Serve()
}
