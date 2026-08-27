package zarinpal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

// Code identifies an error returned by the Zarinpal API.
type Code string

const (
	// CodeValidationFailed indicates invalid request data.
	CodeValidationFailed Code = "-9"
	// CodeRequestNotFound indicates that the requested payment does not exist.
	CodeRequestNotFound Code = "-11"
	// CodeNoFinancialOperation indicates that the transaction has no financial operation.
	CodeNoFinancialOperation Code = "-21"
	// CodeTransactionFailed indicates that the transaction was unsuccessful.
	CodeTransactionFailed Code = "-22"
	// CodeAmountMismatch indicates that the requested and paid amounts differ.
	CodeAmountMismatch Code = "-33"
	// CodeAccessDenied indicates that the merchant cannot use the requested operation.
	CodeAccessDenied Code = "-40"
	// CodeInvalidMetadata indicates invalid additional request data.
	CodeInvalidMetadata Code = "-41"
	// CodeRequestArchived indicates that the payment request was archived.
	CodeRequestArchived Code = "-54"
)

// Error reports a failure returned by Zarinpal or its transport.
type Error struct {
	Operation string
	Code      Code
	Message   string
	Cause     error
	kind      error
}

func (e *Error) Error() string {
	detail := strings.TrimSpace(e.Message)
	if detail == "" && e.Cause != nil {
		detail = e.Cause.Error()
	}
	if e.Code != "" {
		if detail == "" {
			detail = "code " + string(e.Code)
		} else {
			detail = fmt.Sprintf("code %s: %s", e.Code, detail)
		}
	}
	if detail == "" {
		detail = "failed"
	}
	return "zarinpal " + e.Operation + ": " + detail
}

// Unwrap exposes the portable error category and underlying cause.
func (e *Error) Unwrap() []error {
	unwrapped := make([]error, 0, 2)
	if e.kind != nil {
		unwrapped = append(unwrapped, e.kind)
	}
	if e.Cause != nil {
		unwrapped = append(unwrapped, e.Cause)
	}
	return unwrapped
}

// CodeOf returns the Zarinpal error code carried by err.
func CodeOf(err error) (Code, bool) {
	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Code == "" {
		return "", false
	}
	return providerError.Code, true
}

func newError(operation string, code int, message string, cause error) *Error {
	providerCode := Code("")
	if code != 0 {
		providerCode = Code(strconv.Itoa(code))
	}
	return &Error{
		Operation: operation,
		Code:      providerCode,
		Message:   message,
		Cause:     cause,
		kind:      errorKind(providerCode, cause),
	}
}

func errorKind(code Code, cause error) error {
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
