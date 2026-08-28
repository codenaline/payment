package sep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codenaline/payment"
)

func TestPostRejectsHTTPErrorBeforeDecoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"status":1}`, http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	gateway := &Gateway{client: server.Client()}
	var result struct{ Status int }
	err := gateway.post(context.Background(), server.URL, struct{}{}, &result)
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("post() error = %v, want ErrProvider", err)
	}
}

func TestPostRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxResponseSize+1)))
	}))
	t.Cleanup(server.Close)

	gateway := &Gateway{client: server.Client()}
	var result struct{}
	err := gateway.post(context.Background(), server.URL, struct{}{}, &result)
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("post() error = %v, want ErrProvider", err)
	}
}

func TestPostPreservesDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	t.Cleanup(server.Close)

	gateway := &Gateway{client: server.Client()}
	var result struct{}
	err := gateway.post(context.Background(), server.URL, struct{}{}, &result)
	if !errors.Is(err, payment.ErrProvider) || !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("post() error = %v, want decode ErrProvider", err)
	}
}

func TestPostClassifiesTransportFailure(t *testing.T) {
	gateway := &Gateway{client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection failed")
	})}}
	var result struct{}
	err := gateway.post(context.Background(), "https://sep.invalid", struct{}{}, &result)
	if !errors.Is(err, payment.ErrNetwork) {
		t.Fatalf("post() error = %v, want ErrNetwork", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
