package payment

import (
	"errors"
	"strings"
)

// Error describes a payment operation failure while preserving both its
// portable category and underlying cause.
type Error struct {
	Provider  string
	Operation string
	Code      string
	Message   string
	Kind      error
	Cause     error
}

func (e *Error) Error() string {
	prefix := strings.TrimSpace(strings.Join([]string{e.Provider, e.Operation}, " "))
	if prefix == "" {
		prefix = "payment"
	}

	detail := strings.TrimSpace(e.Message)
	if detail == "" && e.Cause != nil {
		detail = e.Cause.Error()
	}
	if e.Code != "" {
		if detail == "" {
			detail = "code " + e.Code
		} else {
			detail = "code " + e.Code + ": " + detail
		}
	}
	if detail == "" {
		detail = "failed"
	}

	return prefix + ": " + detail
}

// Unwrap exposes both the portable payment error and the underlying cause.
func (e *Error) Unwrap() []error {
	unwrapped := make([]error, 0, 2)
	if e.Kind != nil {
		unwrapped = append(unwrapped, e.Kind)
	}
	if e.Cause != nil {
		unwrapped = append(unwrapped, e.Cause)
	}
	return unwrapped
}

// Sentinel errors returned by gateways. Gateway implementations should wrap
// these errors with provider-specific context so callers can inspect them with
// errors.Is.
var (
	ErrInvalidRequest      = errors.New("payment: invalid request")
	ErrPurchaseFailed      = errors.New("payment: purchase failed")
	ErrVerificationFailed  = errors.New("payment: verification failed")
	ErrRefundFailed        = errors.New("payment: refund failed")
	ErrTransactionNotFound = errors.New("payment: transaction not found")
	ErrUnsupported         = errors.New("payment: operation unsupported")
)
