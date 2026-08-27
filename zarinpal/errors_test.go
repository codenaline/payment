package zarinpal

import (
	"errors"
	"testing"

	"github.com/codenaline/payment"
)

func TestCodeOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code Code
		ok   bool
	}{
		{
			name: "Zarinpal error",
			err:  &payment.Error{Provider: providerName, Code: string(CodeAmountMismatch)},
			code: CodeAmountMismatch,
			ok:   true,
		},
		{
			name: "wrapped Zarinpal error",
			err:  errors.Join(errors.New("context"), &payment.Error{Provider: providerName, Code: string(CodeRequestArchived)}),
			code: CodeRequestArchived,
			ok:   true,
		},
		{
			name: "different provider",
			err:  &payment.Error{Provider: "nextpay", Code: "-9"},
		},
		{
			name: "ordinary error",
			err:  errors.New("failure"),
		},
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
