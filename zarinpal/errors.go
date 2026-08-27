package zarinpal

import (
	"fmt"
	"strings"
)

// Error reports a failure returned by Zarinpal or its transport.
type Error struct {
	Operation string
	Code      int
	Message   string
	Kind      error
	Cause     error
}

func (e *Error) Error() string {
	detail := strings.TrimSpace(e.Message)
	if detail == "" && e.Cause != nil {
		detail = e.Cause.Error()
	}
	if e.Code != 0 {
		detail = fmt.Sprintf("code %d: %s", e.Code, detail)
	}
	if detail == "" {
		return "zarinpal " + e.Operation + " failed"
	}
	return "zarinpal " + e.Operation + ": " + detail
}

// Unwrap exposes both the portable payment error and the underlying cause.
func (e *Error) Unwrap() []error {
	errors := make([]error, 0, 2)
	if e.Kind != nil {
		errors = append(errors, e.Kind)
	}
	if e.Cause != nil {
		errors = append(errors, e.Cause)
	}
	return errors
}
