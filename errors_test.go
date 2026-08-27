package payment_test

import (
	"errors"
	"io"
	"testing"

	"github.com/codenaline/payment"
)

func TestError(t *testing.T) {
	t.Parallel()

	err := &payment.Error{
		Provider:  "provider",
		Operation: "purchase",
		Code:      "declined",
		Message:   "payment was declined",
		Kind:      payment.ErrPurchaseFailed,
		Cause:     io.ErrUnexpectedEOF,
	}

	if got, want := err.Error(), "provider purchase: code declined: payment was declined"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, payment.ErrPurchaseFailed) {
		t.Fatal("Error does not unwrap its portable kind")
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal("Error does not unwrap its underlying cause")
	}
}
