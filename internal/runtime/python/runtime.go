package python

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	run "github.com/polyxia-org/agent/internal/runtime"
	log "github.com/sirupsen/logrus"
)

const (
	runtimeName = "python"
)

type runtime struct {
	Logger *log.Entry
}

// Ensure implementation of `Runtime` interface
var _ run.Runtime = (*runtime)(nil)

func New() (*runtime, error) {
	r := &runtime{}

	// Check for runtime version on the host.
	// If an error occurs here, it potentially means that
	// the underlying tool isn't installed or can't be found.
	version, err := r.Version()
	if err != nil {
		return nil, err
	}

	r.Logger = log.New().
		WithField("runtime", r.Name()).
		WithField("version", version)

	r.Logger.Info("Runtime initialized")

	return r, nil
}

// Name return the name of the current runtime.
func (r *runtime) Name() string {
	return runtimeName
}

// Version retrieve the version of the current runtime on host.
// An error can be returned if the executable can't be found in $PATH,
// or if the command can't be executed for any reasons.
func (r *runtime) Version() (string, error) {
	cmd := exec.Command("python", "-c", "import sys; print(sys.version[:6])")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Execute runs a function using the runtime.
func (r *runtime) Execute(iid, wd string, vars run.FnVars) (*run.FnExecution, error) {
	logger := r.Logger.WithField("wd", wd).WithField("iid", iid)

	// First, we need to check for a requirements.txt file inside the current working directory
	// and if it exists, we run the dependencies installation task
	if _, err := os.Stat(filepath.Join(wd, "requirements.txt")); !os.IsNotExist(err) {
		if err := r.installDependencies(wd); err != nil {
			return nil, err
		}
	}

	logger.Info("Injecting custom function wrapper")
	// We assume that an index.js file is present into the working directory.
	// The index.js file must export a function called `handler` in order to be executed.
	// We need to inject a custom wrapper in order to pass context / variables to our functions.
	wrapper := []byte(`
from main import handler
import sys, json

context = json.loads(sys.argv[1]);

print(handler(context))
	`)

	trigger := fmt.Sprintf("%s.py", iid)
	if err := os.WriteFile(filepath.Join(wd, trigger), wrapper, 0644); err != nil {
		panic(err)
	}

	jsonVars, _ := json.Marshal(vars)

	cmd := exec.Command("python", trigger, string(jsonVars))
	cmd.Dir = wd
	out, _ := cmd.CombinedOutput()

	return &run.FnExecution{
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   string(out),
	}, nil
}

func (r *runtime) installDependencies(wd string) error {
	r.Logger.Info("Installing dependencies")
	cmd := exec.Command("pip", "install", "-r", "requirements.txt")
	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	r.Logger.Trace(string(out))
	r.Logger.Info("Dependencies installed")
	return err
}
