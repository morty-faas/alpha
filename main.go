package main

import (
	httpapi "github.com/polyxia-org/agent/internal/api/http"
	"github.com/polyxia-org/agent/internal/executor"
	log "github.com/sirupsen/logrus"
)

func main() {
	exec, err := executor.New()
	if err != nil {
		log.Errorf("failed to initialize executor : %v", exec)
	}

	httpapi.
		NewServer(exec).
		Serve()
}
