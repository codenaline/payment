package zarinpal

import (
	"errors"
	"net/http"
	"testing"

	"github.com/codenaline/payment"
)

func TestNewRequiresMerchantID(t *testing.T) {
	t.Parallel()

	_, err := New(" ")
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("New() error = %v, want ErrInvalidRequest", err)
	}
}

func TestNewAppliesOptions(t *testing.T) {
	t.Parallel()

	client := &http.Client{}
	gateway, err := New("merchant", WithHTTPClient(client), WithSandbox())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client != client {
		t.Fatal("New() did not use the configured HTTP client")
	}
	if gateway.apiBaseURL != sandboxAPIBaseURL || gateway.payBaseURL != sandboxPayBaseURL {
		t.Fatal("New() did not configure sandbox endpoints")
	}
}

func TestWithHTTPClientRejectsNil(t *testing.T) {
	t.Parallel()

	_, err := New("merchant", WithHTTPClient(nil))
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("New() error = %v, want ErrInvalidRequest", err)
	}
}
