package runtime

// Runtime is a generic interface that each runtime we want to add must implement.
type Runtime interface {
	Name() string
	Version() (string, error)
	Execute(functionId, workingDirectory string, vars FnVars) (*FnExecution, error)
}

// FnVars is an alias to a map type used to pass variables to our function invocation.
type FnVars = map[string]interface{}

// FnExecution hold data about a function execution by a runtime.
type FnExecution struct {
	Output   string `json:"output"`
	ExitCode int    `json:"process_exit_code"`
}
