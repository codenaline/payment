package zarinpal

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/codenaline/payment"
)

func TestNewRequiresMerchantID(t *testing.T) {
	t.Parallel()

	_, err := New(Config{MerchantID: " "})
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("New() error = %v, want ErrInvalidRequest", err)
	}
}

func TestNewAppliesConfig(t *testing.T) {
	t.Parallel()

	client := &http.Client{}
	gateway, err := New(Config{
		MerchantID: "merchant",
		Sandbox:    true,
		HTTPClient: client,
	})
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

func TestNewProvidesDefaultHTTPClient(t *testing.T) {
	t.Parallel()

	gateway, err := New(Config{MerchantID: "merchant"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client == nil || gateway.client.Timeout != 30*time.Second {
		t.Fatalf("New() HTTP client = %#v, want 30 second timeout", gateway.client)
	}
}
