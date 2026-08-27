package nextpay

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/codenaline/payment"
)

func TestPostRejectsHTTPErrorBeforeDecoding(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("not JSON"))
	}))
	defer server.Close()
	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL
	err := gateway.post(t.Context(), "/", url.Values{}, &struct{}{})
	if !errors.Is(err, payment.ErrProvider) || !strings.Contains(err.Error(), "502 Bad Gateway") {
		t.Fatalf("post() error = %v, want HTTP status error", err)
	}
}

func TestPostRejectsOversizedResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxResponseSize+1)))
	}))
	defer server.Close()
	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL
	err := gateway.post(t.Context(), "/", url.Values{}, &struct{}{})
	if !errors.Is(err, payment.ErrProvider) || !strings.Contains(err.Error(), "response exceeds") {
		t.Fatalf("post() error = %v, want response size error", err)
	}
}

func TestPostPreservesDecodeError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer server.Close()
	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL
	err := gateway.post(t.Context(), "/", url.Values{}, &struct{}{})
	var syntaxError *json.SyntaxError
	if !errors.Is(err, payment.ErrProvider) || !errors.As(err, &syntaxError) {
		t.Fatalf("post() error = %v, want wrapped JSON syntax error", err)
	}
}
