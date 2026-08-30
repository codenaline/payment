package aqayepardakht

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/codenaline/payment"
)

func TestPostFormRejectsHTTPErrorBeforeDecoding(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "pin", HTTPClient: server.Client()})
	var response apiResponse
	err := gateway.postForm(t.Context(), server.URL, url.Values{}, &response)
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("postForm() error = %v, want ErrProvider", err)
	}
}

func TestPostFormRejectsOversizedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxResponseSize+1)))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "pin", HTTPClient: server.Client()})
	var response apiResponse
	err := gateway.postForm(t.Context(), server.URL, url.Values{}, &response)
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("postForm() error = %v, want ErrProvider", err)
	}
}

func TestPostFormClassifiesNetworkFailure(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})}
	gateway, _ := New(Config{Pin: "pin", HTTPClient: client})
	var response apiResponse
	err := gateway.postForm(t.Context(), "https://provider.example", url.Values{}, &response)
	if !errors.Is(err, payment.ErrNetwork) {
		t.Fatalf("postForm() error = %v, want ErrNetwork", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
