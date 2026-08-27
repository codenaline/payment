package zarinpal

import (
	"errors"
	"io"
	"testing"

	"github.com/codenaline/payment"
)

func TestErrorSupportsPortableAndProviderInspection(t *testing.T) {
	t.Parallel()

	err := newError("verification", -54, "request archived", io.ErrUnexpectedEOF)
	if !errors.Is(err, payment.ErrTransactionNotFound) {
		t.Fatal("Error does not unwrap its portable category")
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal("Error does not unwrap its underlying cause")
	}

	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Code != CodeRequestArchived {
		t.Fatalf("errors.As() = %#v, want CodeRequestArchived", providerError)
	}
}
