package node

import (
	"encoding/json"
	"fmt"
	run "github.com/polyxia-org/agent/internal/runtime"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	runtimeName = "node"
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
	cmd := exec.Command("node", "-v")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

// Execute runs a function using the runtime.
func (r *runtime) Execute(iid, wd string, vars run.FnVars) (*run.FnExecution, error) {
	logger := r.Logger.WithField("wd", wd).WithField("iid", iid)

	// First, we need to check for a package.json file inside the current working directory
	// and if it exists, we run the dependencies installation task
	if _, err := os.Stat(filepath.Join(wd, "package.json")); !os.IsNotExist(err) {
		if err := r.installDependencies(wd); err != nil {
			return nil, err
		}
	}

	logger.Info("Injecting custom function wrapper")
	// We assume that an index.js file is present into the working directory.
	// The index.js file must export a function called `handler` in order to be executed.
	// We need to inject a custom wrapper in order to pass context / variables to our functions.
	wrapper := []byte(`
		const fn = require("./index")
		
		const context = JSON.parse(process.argv[2]);
		
		console.log(fn.handler(context))
	`)

	trigger := fmt.Sprintf("%s.js", iid)
	if err := os.WriteFile(filepath.Join(wd, trigger), wrapper, 0644); err != nil {
		panic(err)
	}

	jsonVars, _ := json.Marshal(vars)

	cmd := exec.Command("node", trigger, string(jsonVars))
	cmd.Dir = wd
	out, _ := cmd.CombinedOutput()

	return &run.FnExecution{
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   string(out),
	}, nil
}

func (r *runtime) installDependencies(wd string) error {
	cmd := exec.Command("npm", "install")
	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	r.Logger.Trace(string(out))
	return err
}
