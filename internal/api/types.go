package api

import (
	"errors"
	"net/url"
)

type (
	// InvokeRequest is the body expected to be received from an invocation request.
	InvokeRequest struct {
		Runtime string `json:"runtime"`
		Code    string `json:"code"`
	}

	HealthResponse struct {
		Status string `json:"status"`
	}
)

var (
	ErrCodeUrlEmpty   = errors.New("invokeRequest/validate: `code` can't be empty")
	ErrCodeUrlInvalid = errors.New("invokeRequest/validate: `code` isn't a valid URL")
)

func (ir *InvokeRequest) Validate() error {
	if ir.Code == "" {
		return ErrCodeUrlEmpty
	}

	if _, err := url.Parse(ir.Code); err != nil {
		return ErrCodeUrlInvalid
	}

	return nil
}
