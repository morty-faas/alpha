package main

type (
	InstrumentedResponse struct {
		Payload any              `json:"payload"`
		Process *ProcessMetadata `json:"process_metadata"`
	}

	ProcessMetadata struct {
		State           string   `json:"state"`
		ExecutionTimeMs int64    `json:"execution_time_ms"`
		Logs            []string `json:"logs"`
	}

	HealthResponse struct {
		Status string `json:"status"`
	}
)
