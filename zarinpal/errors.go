package zarinpal

import (
	"errors"

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

// CodeOf returns the Zarinpal error code carried by err.
func CodeOf(err error) (Code, bool) {
	var paymentError *payment.Error
	if !errors.As(err, &paymentError) || paymentError.Provider != providerName || paymentError.Code == "" {
		return "", false
	}
	return Code(paymentError.Code), true
}
