package runtime

import (
	"context"
	"os/exec"
)

// Runtime is a generic interface that each runtime we want to add must implement.
type Runtime interface {
	Name() string
	Version() (string, error)
	WrapCmd(ctx context.Context) (*exec.Cmd, error)
}

// FnParams is an alias to a map type used to pass parameters to our function invocation.
type FnParams = map[string]interface{}

// FnExecution hold data about a function execution by a runtime.
type FnExecution struct {
	Output   string   `json:"output"`
	Logs     []string `json:"logs"`
	ExitCode int      `json:"process_exit_code"`
}
