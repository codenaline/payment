package nextpay

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/codenaline/payment"
)

func TestNewRequiresAPIKey(t *testing.T) {
	t.Parallel()

	_, err := New(Config{APIKey: " "})
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("New() error = %v, want ErrInvalidRequest", err)
	}
}

func TestNewUsesConfiguredHTTPClient(t *testing.T) {
	t.Parallel()

	client := &http.Client{}
	gateway, err := New(Config{APIKey: "key", HTTPClient: client})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client != client {
		t.Fatal("New() did not use the configured HTTP client")
	}
}

func TestNewProvidesDefaultHTTPClient(t *testing.T) {
	t.Parallel()

	gateway, err := New(Config{APIKey: "key"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client == nil || gateway.client.Timeout != 30*time.Second {
		t.Fatalf("New() HTTP client = %#v, want 30 second timeout", gateway.client)
	}
}
