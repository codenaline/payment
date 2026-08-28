package sep

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codenaline/payment"
)

// Code identifies a result returned by SEP.
type Code int

const (
	// CodeSuccess indicates a successful operation.
	CodeSuccess Code = 0
	// CodeTransactionNotFound indicates that the requested receipt does not exist.
	CodeTransactionNotFound Code = -2
	// CodeVerificationExpired indicates that the verification window has elapsed.
	CodeVerificationExpired Code = -6
	// CodeDuplicateRequest indicates a repeated verify or reverse request.
	CodeDuplicateRequest Code = 2
	// CodeInvalidParameters indicates invalid token-request parameters. During
	// verification, the same numeric code indicates a reversed transaction.
	CodeInvalidParameters Code = 5
	// CodeTerminalInactive indicates an inactive terminal.
	CodeTerminalInactive Code = -104
	// CodeTerminalNotFound indicates that the terminal does not exist.
	CodeTerminalNotFound Code = -105
	// CodeInvalidIPAddress indicates that the merchant server IP is not allowed.
	CodeInvalidIPAddress Code = -106
)

// Error represents an error returned while communicating with SEP.
type Error struct {
	Operation string
	Code      Code
	Message   string
	Cause     error
}

func (e *Error) Error() string {
	detail := strings.TrimSpace(e.Message)
	if detail == "" && e.Cause != nil {
		detail = e.Cause.Error()
	}
	if e.Code != 0 {
		if detail == "" {
			detail = fmt.Sprintf("code %d", e.Code)
		} else {
			detail = fmt.Sprintf("code %d: %s", e.Code, detail)
		}
	}
	if detail == "" {
		detail = "failed"
	}
	return "sep " + e.Operation + ": " + detail
}

// Is reports whether the error belongs to a portable payment error category.
func (e *Error) Is(target error) bool {
	return errors.Is(classify(e.Operation, e.Code, e.Cause), target)
}

// Unwrap returns the underlying cause, if any.
func (e *Error) Unwrap() error {
	return e.Cause
}

func newError(operation string, code int, message string, cause error) *Error {
	return &Error{
		Operation: operation,
		Code:      Code(code),
		Message:   message,
		Cause:     cause,
	}
}

func classify(operation string, code Code, cause error) error {
	switch {
	case errors.Is(cause, payment.ErrNetwork):
		return payment.ErrNetwork
	case errors.Is(cause, payment.ErrInvalidRequest):
		return payment.ErrInvalidRequest
	case errors.Is(cause, payment.ErrTransactionNotFound):
		return payment.ErrTransactionNotFound
	case errors.Is(cause, payment.ErrCanceled):
		return payment.ErrCanceled
	case errors.Is(cause, payment.ErrProvider):
		return payment.ErrProvider
	}

	if operation == "purchase" && code == CodeInvalidParameters {
		return payment.ErrInvalidRequest
	}
	if operation == "verification" {
		switch code {
		case CodeTransactionNotFound, CodeVerificationExpired:
			return payment.ErrTransactionNotFound
		case CodeInvalidParameters:
			return payment.ErrCanceled
		case CodeTerminalNotFound:
			return payment.ErrInvalidRequest
		}
	}
	return payment.ErrProvider
}
