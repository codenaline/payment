package payment

import "errors"

// Portable error categories returned by payment gateways. Gateway
// implementations may wrap these errors with provider-specific details.
var (
	ErrInvalidRequest      = errors.New("payment: invalid request")
	ErrNetwork             = errors.New("payment: network error")
	ErrTransactionNotFound = errors.New("payment: transaction not found")
	ErrCanceled            = errors.New("payment: canceled")
	ErrDeclined            = errors.New("payment: declined")
	ErrProvider            = errors.New("payment: provider error")
	ErrUnsupported         = errors.New("payment: operation unsupported")
)
