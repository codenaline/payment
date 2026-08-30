package aqayepardakht

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/codenaline/payment"
)

func TestNewRejectsBlankPIN(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Pin: "  "})
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("New() error = %v, want ErrInvalidRequest", err)
	}
}

func TestNewUsesSafeDefaultHTTPClient(t *testing.T) {
	t.Parallel()

	gateway, err := New(Config{Pin: "pin"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client.Timeout != 30*time.Second {
		t.Fatalf("HTTP timeout = %v, want 30s", gateway.client.Timeout)
	}
}

func TestNewPreservesCustomHTTPClient(t *testing.T) {
	t.Parallel()

	client := &http.Client{Timeout: time.Second}
	gateway, err := New(Config{Pin: "pin", HTTPClient: client})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gateway.client != client {
		t.Fatal("New() did not preserve custom HTTP client")
	}
}
