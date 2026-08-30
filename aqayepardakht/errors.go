package aqayepardakht

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codenaline/payment"
)

// Code identifies a result returned by Aghaye Pardakht.
type Code int

const (
	CodePaymentFailed       Code = 0
	CodeSuccess             Code = 1
	CodeAlreadyVerified     Code = 2
	CodeAmountRequired      Code = -1
	CodePINRequired         Code = -2
	CodeCallbackRequired    Code = -3
	CodeInvalidAmount       Code = -4
	CodeAmountOutOfRange    Code = -5
	CodeInvalidPIN          Code = -6
	CodeTransactionRequired Code = -7
	CodeTransactionNotFound Code = -8
	CodePINMismatch         Code = -9
	CodeAmountMismatch      Code = -10
	CodeGatewayInactive     Code = -11
	CodeMerchantRejected    Code = -12
	CodeInvalidCardNumber   Code = -13
)

// Error represents an error returned while communicating with Aghaye Pardakht.
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
	return "aqayepardakht " + e.Operation + ": " + detail
}

// Is reports whether the error belongs to a portable payment error category.
func (e *Error) Is(target error) bool {
	return errors.Is(classify(e.Operation, e.Code, e.Cause), target)
}

// Unwrap returns the underlying cause, if any.
func (e *Error) Unwrap() error {
	return e.Cause
}

func newError(operation string, code Code, message string, cause error) *Error {
	return &Error{Operation: operation, Code: code, Message: message, Cause: cause}
}

func classify(operation string, code Code, cause error) error {
	switch {
	case errors.Is(cause, payment.ErrNetwork):
		return payment.ErrNetwork
	case errors.Is(cause, payment.ErrInvalidRequest):
		return payment.ErrInvalidRequest
	case errors.Is(cause, payment.ErrProvider):
		return payment.ErrProvider
	}

	switch code {
	case CodeAmountRequired, CodePINRequired, CodeCallbackRequired, CodeInvalidAmount,
		CodeAmountOutOfRange, CodeInvalidPIN, CodePINMismatch, CodeAmountMismatch,
		CodeInvalidCardNumber:
		return payment.ErrInvalidRequest
	case CodeTransactionRequired, CodeTransactionNotFound:
		return payment.ErrTransactionNotFound
	}
	if operation == "verification" && code == CodePaymentFailed {
		return payment.ErrDeclined
	}
	return payment.ErrProvider
}

func message(code Code) string {
	switch code {
	case CodePaymentFailed:
		return "payment was not completed"
	case CodeAlreadyVerified:
		return "transaction was already verified"
	case CodeAmountRequired:
		return "amount is required"
	case CodePINRequired:
		return "gateway PIN is required"
	case CodeCallbackRequired:
		return "callback is required"
	case CodeInvalidAmount:
		return "amount must be numeric"
	case CodeAmountOutOfRange:
		return "amount must be between 100 and 50000000 toman"
	case CodeInvalidPIN:
		return "gateway PIN is invalid"
	case CodeTransactionRequired:
		return "transaction ID is required"
	case CodeTransactionNotFound:
		return "transaction was not found"
	case CodePINMismatch:
		return "gateway PIN does not match transaction"
	case CodeAmountMismatch:
		return "amount does not match transaction"
	case CodeGatewayInactive:
		return "gateway is inactive or awaiting approval"
	case CodeMerchantRejected:
		return "merchant cannot submit requests"
	case CodeInvalidCardNumber:
		return "card number must contain 16 digits"
	default:
		return "Aghaye Pardakht request failed"
	}
}
