package aqayepardakht

import (
	"errors"
	"testing"

	"github.com/codenaline/payment"
)

func TestErrorPreservesPortableCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("connection reset")
	err := newError("purchase", 0, "request failed", errors.Join(payment.ErrNetwork, cause))
	if !errors.Is(err, payment.ErrNetwork) {
		t.Fatalf("error = %v, want ErrNetwork", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("error = %v, want wrapped cause", err)
	}
}

func TestUnknownCodeIsProviderError(t *testing.T) {
	t.Parallel()

	err := newError("verification", Code(-999), "unknown", nil)
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("error = %v, want ErrProvider", err)
	}
}
