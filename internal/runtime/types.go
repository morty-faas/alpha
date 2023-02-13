package runtime

type (
	FnInvocation struct {
		CodeURL string
		Runtime string
		Params  FnParams
	}

	FnExecutionResult struct {
		Payload     interface{}             `json:"payload"`
		ProcessInfo *FnExecutionProcessInfo `json:"process"`
	}

	FnExecutionProcessInfo struct {
		RuntimeInfo   *FnExecutionRuntimeInfo `json:"runtime"`
		ExecutionTime int64                   `json:"execution_time_millis"`
		ExitCode      int                     `json:"exit_code"`
		Logs          []string                `json:"logs"`
	}
	FnExecutionRuntimeInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
)
