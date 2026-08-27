package zarinpal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codenaline/payment"
)

// Code identifies an error returned by Zarinpal.
type Code int

const (
	// CodeValidationFailed indicates invalid request data.
	CodeValidationFailed Code = -9
	// CodeRequestNotFound indicates that the requested payment does not exist.
	CodeRequestNotFound Code = -11
	// CodeNoFinancialOperation indicates that the transaction has no financial operation.
	CodeNoFinancialOperation Code = -21
	// CodeTransactionFailed indicates that the transaction was unsuccessful.
	CodeTransactionFailed Code = -22
	// CodeAmountMismatch indicates that the requested and paid amounts differ.
	CodeAmountMismatch Code = -33
	// CodeAccessDenied indicates that the merchant cannot use the requested operation.
	CodeAccessDenied Code = -40
	// CodeInvalidMetadata indicates invalid additional request data.
	CodeInvalidMetadata Code = -41
	// CodeRequestArchived indicates that the payment request was archived.
	CodeRequestArchived Code = -54
)

// Error represents an error returned by Zarinpal.
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
	return "zarinpal " + e.Operation + ": " + detail
}

// Is reports whether the error belongs to a portable payment error category.
func (e *Error) Is(target error) bool {
	return errors.Is(classify(e.Code, e.Cause), target)
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

func classify(code Code, cause error) error {
	switch {
	case errors.Is(cause, payment.ErrNetwork):
		return payment.ErrNetwork
	case errors.Is(cause, payment.ErrInvalidRequest):
		return payment.ErrInvalidRequest
	}

	switch code {
	case CodeValidationFailed, CodeAmountMismatch, CodeInvalidMetadata:
		return payment.ErrInvalidRequest
	case CodeRequestNotFound, CodeNoFinancialOperation, CodeRequestArchived:
		return payment.ErrTransactionNotFound
	case CodeTransactionFailed:
		return payment.ErrDeclined
	default:
		return payment.ErrProvider
	}
}
