package zarinpal

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codenaline/payment"
)

func TestPostRejectsHTTPErrorBeforeDecoding(t *testing.T) {
	t.Parallel()
	gateway, request := testRequest(t, http.StatusBadGateway, "not JSON")
	err := gateway.post(request, &struct{}{})
	if !errors.Is(err, payment.ErrProvider) || !strings.Contains(err.Error(), "502 Bad Gateway") {
		t.Fatalf("post() error = %v, want HTTP status error", err)
	}
}

func TestPostRejectsOversizedResponse(t *testing.T) {
	t.Parallel()
	gateway, request := testRequest(t, http.StatusOK, strings.Repeat("x", maxResponseSize+1))
	err := gateway.post(request, &struct{}{})
	if !errors.Is(err, payment.ErrProvider) || !strings.Contains(err.Error(), "response exceeds") {
		t.Fatalf("post() error = %v, want response size error", err)
	}
}

func TestPostPreservesDecodeError(t *testing.T) {
	t.Parallel()
	gateway, request := testRequest(t, http.StatusOK, "{")
	err := gateway.post(request, &struct{}{})
	var syntaxError *json.SyntaxError
	if !errors.Is(err, payment.ErrProvider) || !errors.As(err, &syntaxError) {
		t.Fatalf("post() error = %v, want wrapped JSON syntax error", err)
	}
}

func testRequest(t *testing.T, status int, body string) (*Gateway, *http.Request) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	gateway, err := New(Config{MerchantID: "merchant", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	return gateway, request
}
