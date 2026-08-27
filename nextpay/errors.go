package nextpay

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codenaline/payment"
)

// Code identifies a status returned by NextPay.
type Code int

const (
	// CodeTokenCreated indicates successful token creation.
	CodeTokenCreated Code = -1
	// CodePaymentRejected indicates rejection by the payer or bank.
	CodePaymentRejected Code = -2
	// CodePaymentPending indicates that NextPay is waiting for the bank.
	CodePaymentPending Code = -3
	// CodePaymentCanceled indicates a canceled payment.
	CodePaymentCanceled Code = -4
	// CodeInvalidAmount indicates an invalid amount.
	CodeInvalidAmount Code = -24
	// CodeInvalidAPIKey indicates invalid API credentials.
	CodeInvalidAPIKey Code = -33
	// CodeInvalidTransaction indicates an invalid transaction token.
	CodeInvalidTransaction Code = -34
	// CodeTransactionNotFound indicates that the transaction does not exist.
	CodeTransactionNotFound Code = -37
	// CodeSystemError indicates a NextPay system failure.
	CodeSystemError Code = -42
)

// Error represents an error returned by NextPay.
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
	return "nextpay " + e.Operation + ": " + detail
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
	return &Error{Operation: operation, Code: Code(code), Message: message, Cause: cause}
}

func classify(code Code, cause error) error {
	switch {
	case errors.Is(cause, payment.ErrNetwork):
		return payment.ErrNetwork
	case errors.Is(cause, payment.ErrInvalidRequest):
		return payment.ErrInvalidRequest
	}

	switch code {
	case CodePaymentRejected:
		return payment.ErrDeclined
	case CodePaymentCanceled:
		return payment.ErrCanceled
	case CodeInvalidAmount:
		return payment.ErrInvalidRequest
	case CodeInvalidTransaction, CodeTransactionNotFound:
		return payment.ErrTransactionNotFound
	default:
		return payment.ErrProvider
	}
}
