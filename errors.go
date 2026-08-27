package payment

import "errors"

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
