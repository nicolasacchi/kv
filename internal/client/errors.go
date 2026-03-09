package client

import (
	"fmt"
	"net/http"
)

// APIError represents a structured error from the Klaviyo API.
type APIError struct {
	StatusCode int
	Code       string
	Title      string
	Detail     string
	Hint       string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Detail)
	}
	if e.Title != "" {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Title)
	}
	return fmt.Sprintf("%d: %s", e.StatusCode, http.StatusText(e.StatusCode))
}

func (e *APIError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return 3 // auth error
	case e.StatusCode == 404:
		return 1 // API error (not found)
	default:
		return 1 // generic API error
	}
}
