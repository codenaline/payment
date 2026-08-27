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

func TestCodeOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code Code
		ok   bool
	}{
		{name: "Zarinpal error", err: newError("verify", -33, "", nil), code: CodeAmountMismatch, ok: true},
		{name: "wrapped Zarinpal error", err: errors.Join(errors.New("context"), newError("verify", -54, "", nil)), code: CodeRequestArchived, ok: true},
		{name: "ordinary error", err: errors.New("failure")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, ok := CodeOf(test.err)
			if code != test.code || ok != test.ok {
				t.Fatalf("CodeOf() = %q, %t; want %q, %t", code, ok, test.code, test.ok)
			}
		})
	}
}
