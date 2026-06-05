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

	// Kind classifies client-side guard errors raised before a request
	// (StatusCode stays 0). Currently: "write_locked" (confirm gate).
	Kind string
}

func (e *APIError) Error() string {
	// Client-side guard errors (no HTTP status) render from Detail/Hint.
	if e.Kind != "" && e.StatusCode == 0 {
		msg := e.Detail
		if msg == "" {
			msg = e.Kind
		}
		if e.Hint != "" {
			return fmt.Sprintf("%s — %s", msg, e.Hint)
		}
		return msg
	}
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
	case e.Kind == "write_locked":
		return 6 // refused for safety, not failed — matches the otx/stx write-gate contract
	case e.StatusCode == 401 || e.StatusCode == 403:
		return 3 // auth error
	case e.StatusCode == 404:
		return 1 // API error (not found)
	default:
		return 1 // generic API error
	}
}
