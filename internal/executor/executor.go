package executor

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/polyxia-org/agent/internal/api"
	"github.com/polyxia-org/agent/internal/runtime"
	"github.com/polyxia-org/agent/internal/runtime/node"
	"github.com/polyxia-org/agent/internal/runtime/python"
	"github.com/polyxia-org/agent/internal/utils"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	logger   *log.Logger
	runtimes map[string]runtime.Runtime
}

var (
	ErrCodeDownload       = errors.New("executor/invoke: failed to download function code")
	ErrCodeUncompress     = errors.New("executor/invoke: failed to uncompress the function code archive")
	ErrRuntimeUnavailable = errors.New("executor/invoke: runtime unavailable")
)

func New() (*Executor, error) {
	logger := log.New()
	runtimes := map[string]runtime.Runtime{}

	// As for now, we don't have any external configuration, so
	// we initialize all available runtimes. Ideally in a future version,
	// it will be great if we can configure which runtime we want to use
	// on a specific host.
	nodejs, err := node.New()
	if err != nil {
		logger.Errorf("runtime/node failed to initialize: %v", err)
	} else {
		runtimes[nodejs.Name()] = nodejs
	}
	pythonRuntime, err := python.New()
	if err != nil {
		logger.Errorf("runtime/python failed to initialize: %v", err)
	} else {
		runtimes[pythonRuntime.Name()] = pythonRuntime
	}

	logger.Infof("Executor ready. %d runtime(s) available for this agent", len(runtimes))

	return &Executor{logger, runtimes}, nil
}

func (ex *Executor) Invoke(request *api.InvokeRequest, vars runtime.FnVars) (*runtime.FnExecution, error) {
	iid := uuid.New().String()
	logger := ex.logger.WithField("iid", iid)
	logger.Infof("Assigned ID to function invocation")

	logger.Infof("Downloading function code from %s", request.Code)

	name := filepath.Base(request.Code)
	downloadPath := fmt.Sprintf("/tmp/%s", name)
	wd := fmt.Sprintf("/tmp/%s", iid)

	if err := utils.DownloadFile(downloadPath, request.Code); err != nil {
		return nil, ErrCodeDownload
	}

	if err := utils.Untar(downloadPath, wd); err != nil {
		return nil, ErrCodeUncompress
	}

	runtime := ex.runtimes[request.Runtime]
	if runtime == nil {
		return nil, ErrRuntimeUnavailable
	}

	now := time.Now()
	result, err := runtime.Execute(iid, wd, vars)
	elapsed := time.Since(now).Milliseconds()

	logger.Infof("Function exited with exit code %d in %dms", result.ExitCode, elapsed)

	return result, err
}
