package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/polyxia-org/agent/internal/runtime/node"
	"github.com/polyxia-org/agent/internal/runtime/python"
	"github.com/polyxia-org/agent/internal/utils"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
)

type Invoker struct {
	logger   *log.Logger
	runtimes map[string]Runtime
}

var (
	ErrCodeDownload               = errors.New("invoker/invoke: failed to download function code")
	ErrCodeUncompress             = errors.New("invoker/invoke: failed to uncompress the function code archive")
	ErrRuntimeUnavailable         = errors.New("invoker/invoke: runtime unavailable")
	ErrFunctionOutputNotValidJson = errors.New("invoker/invoke: function output isn't a valid JSON")
)

func New() (*Invoker, error) {
	logger := log.New()
	runtimes := map[string]Runtime{}

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

	logger.Infof("Invoker ready. %d runtime(s) available for this agent", len(runtimes))

	return &Invoker{
		logger,
		runtimes,
	}, nil
}

// Invoke try to execute the FnInvocation given in parameters onto one of the registered runtimes.
func (i *Invoker) Invoke(ctx context.Context, fn *FnInvocation) (*FnExecutionResult, error) {
	logger := i.logger.WithField("iid", ctx.Value("iid"))

	// First, we should retrieve the function code from the given
	// URL. If an error occurs during the download process,
	// we should immediately exit as we currently support code that
	// comes from a remote archive.
	wd, err := i.downloadCode(ctx, fn.CodeURL, logger)
	if err != nil {
		return nil, err
	}

	logger = logger.WithField("wd", wd)

	// Retrieve the runtime asked for the function execution.
	// If no runtime matches, we need to return immediately as
	// we will not be able to execute the function.
	runtime := i.runtimes[fn.Runtime]
	if runtime == nil {
		return nil, ErrRuntimeUnavailable
	}

	ctx = context.WithValue(ctx, "wd", wd)

	// Wrap an `exec.Cmd` command with the custom runtime wrapper,
	// and install dependencies or other stuff if required by the runtime.
	cmd, err := runtime.WrapCmd(ctx)
	if err != nil {
		return nil, err
	}

	cmd.Dir = wd

	// Create the parameters map
	params, _ := json.Marshal(fn.Params)
	cmd.Args = append(cmd.Args, string(params))

	// Map stdout and stderr for command.
	// outb will contain the function result, and
	// errb will contain every function execution logs or errors.
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	logger.Debugf("Command : %s", cmd.String())

	// Run the process, and populate outb and errb
	now := time.Now()
	cmd.Run()
	elapsed := time.Since(now).Milliseconds()

	logger.Debugf("'%s' exited with exit code %d in %dms", cmd.String(), cmd.ProcessState.ExitCode(), elapsed)

	logs := strings.Split(strings.TrimSuffix(errb.String(), "\n"), "\n")

	// We can safely skip the error here. If we're inside this function,
	// it means that the runtime was already successfully initialized.
	runtimeVersion, _ := runtime.Version()

	// Parse the function result as JSON. By definition,
	// a function must return a valid JSON object representation.
	var payload map[string]interface{}
	if err := json.Unmarshal(outb.Bytes(), &payload); err != nil {
		log.Error(err)
		return nil, ErrFunctionOutputNotValidJson
	}

	log.Infof("Successfully processed invocation %s in %dms", ctx.Value("iid"), elapsed)

	output := &FnExecutionResult{
		Payload: payload,
		ProcessInfo: &FnExecutionProcessInfo{
			RuntimeInfo: &FnExecutionRuntimeInfo{
				Name:    runtime.Name(),
				Version: runtimeVersion,
			},
			ExitCode:      cmd.ProcessState.ExitCode(),
			ExecutionTime: elapsed,
			Logs:          logs,
		},
	}

	return output, nil
}

// downloadCode download the function code from a remote URL, and tries
// to uncompress the downloaded archive. If the operation is successful, the path
// to the function code will be returned.
func (i *Invoker) downloadCode(ctx context.Context, url string, logger *log.Entry) (string, error) {
	logger.Debugf("Downloading function code from %s", url)

	// Get the name of the file we need to download
	name := filepath.Base(url)

	// Currently, we store all of our working directories
	// into /tmp, for the sake of simplicity.
	downloadPath := fmt.Sprintf("/tmp/%s", name)
	wd := fmt.Sprintf("/tmp/%s", ctx.Value("iid"))

	if err := utils.DownloadFile(downloadPath, url); err != nil {
		return "", ErrCodeDownload
	}

	if err := utils.Untar(downloadPath, wd); err != nil {
		return "", ErrCodeUncompress
	}

	return wd, nil
}
