package sep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/codenaline/payment"
)

func TestPurchaseErrorSupportsPortableAndProviderInspection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":-1,"errorCode":"5","errorDesc":"invalid parameters"}`))
	}))
	t.Cleanup(server.Close)
	gateway, err := New(Config{TerminalID: 12345, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	gateway.tokenURL = server.URL

	_, err = gateway.Purchase(context.Background(), payment.PurchaseRequest{
		OrderID:     "order-42",
		Amount:      payment.Money{Amount: 125_000, Currency: payment.CurrencyIRR},
		CallbackURL: "https://example.com/callback",
	})
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
	}
	var providerError *Error
	if !errors.As(err, &providerError) {
		t.Fatalf("Purchase() error type = %T, want *Error", err)
	}
	if providerError.Operation != "purchase" || providerError.Code != CodeInvalidParameters || providerError.Message != "invalid parameters" {
		t.Fatalf("provider error = %+v", providerError)
	}
}

func TestVerificationErrorClassifiesDocumentedCodes(t *testing.T) {
	tests := map[string]struct {
		code int
		want error
	}{
		"transaction not found": {code: -2, want: payment.ErrTransactionNotFound},
		"verification expired":  {code: -6, want: payment.ErrTransactionNotFound},
		"transaction reversed":  {code: 5, want: payment.ErrCanceled},
		"terminal not found":    {code: -105, want: payment.ErrInvalidRequest},
		"unknown":               {code: -999, want: payment.ErrProvider},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gateway := verificationGateway(t, `{
				"TransactionDetail": {},
				"ResultCode": `+strconv.Itoa(test.code)+`,
				"ResultDescription": "failed",
				"Success": false
			}`)
			_, err := gateway.Verify(context.Background(), validVerifyRequest())
			if !errors.Is(err, test.want) {
				t.Fatalf("Verify() error = %v, want %v", err, test.want)
			}
			var providerError *Error
			if !errors.As(err, &providerError) || int(providerError.Code) != test.code || providerError.Operation != "verification" {
				t.Fatalf("provider error = %#v", providerError)
			}
		})
	}
}

func TestOperationErrorPreservesNetworkCause(t *testing.T) {
	gateway, err := New(Config{
		TerminalID: 12345,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection failed")
		})},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = gateway.Verify(context.Background(), validVerifyRequest())
	if !errors.Is(err, payment.ErrNetwork) {
		t.Fatalf("Verify() error = %v, want ErrNetwork", err)
	}
	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Cause == nil {
		t.Fatalf("provider error = %#v, want cause", providerError)
	}
}
