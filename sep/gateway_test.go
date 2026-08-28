package sep

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/codenaline/payment"
)

func TestNewRequiresPositiveTerminalID(t *testing.T) {
	for _, terminalID := range []int64{0, -1} {
		_, err := New(Config{TerminalID: terminalID})
		if !errors.Is(err, payment.ErrInvalidRequest) {
			t.Fatalf("New() error = %v, want ErrInvalidRequest", err)
		}
	}
}

func TestNewUsesConfiguredHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}

	gateway, err := New(Config{TerminalID: 12345, HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.terminalID != 12345 {
		t.Fatalf("terminalID = %d, want 12345", gateway.terminalID)
	}
	if gateway.client != httpClient {
		t.Fatal("client does not match configured HTTP client")
	}
}

func TestNewProvidesDefaultHTTPClient(t *testing.T) {
	gateway, err := New(Config{TerminalID: 12345})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client == nil {
		t.Fatal("client is nil")
	}
	if gateway.client.Timeout != 30*time.Second {
		t.Fatalf("client timeout = %s, want 30s", gateway.client.Timeout)
	}
}
