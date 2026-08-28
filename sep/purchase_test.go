package sep

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/codenaline/payment"
)

func TestPurchase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		var payload struct {
			Action      string `json:"Action"`
			TerminalID  int64  `json:"TerminalId"`
			Amount      int64  `json:"Amount"`
			ResNum      string `json:"ResNum"`
			RedirectURL string `json:"RedirectUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Action != "Token" || payload.TerminalID != 12345 || payload.Amount != 125_000 || payload.ResNum != "order-42" || payload.RedirectURL != "https://example.com/callback" {
			t.Fatalf("payload = %+v", payload)
		}
		_, _ = w.Write([]byte(`{"status":1,"token":"token/+ value"}`))
	}))
	t.Cleanup(server.Close)

	gateway, err := New(Config{TerminalID: 12345, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	gateway.tokenURL = server.URL
	gateway.payURL = "https://sep.example/OnlinePG/SendToken"

	result, err := gateway.Purchase(context.Background(), payment.PurchaseRequest{
		OrderID: "order-42",
		Amount: payment.Money{
			Amount:   125_000,
			Currency: payment.CurrencyIRR,
		},
		CallbackURL: "https://example.com/callback",
		Description: "not sent by SEP",
	})
	if err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}
	if result.Transaction.ID != "token/+ value" || result.Transaction.Status != payment.StatusPending || result.Transaction.Provider != "sep" || result.Transaction.Amount.Amount != 125_000 || result.Transaction.Amount.Currency != payment.CurrencyIRR {
		t.Fatalf("transaction = %+v", result.Transaction)
	}
	redirect, err := url.Parse(result.RedirectURL)
	if err != nil {
		t.Fatalf("parse redirect URL: %v", err)
	}
	if redirect.Scheme != "https" || redirect.Host != "sep.example" || redirect.Path != "/OnlinePG/SendToken" || redirect.Query().Get("token") != "token/+ value" {
		t.Fatalf("redirect URL = %q", result.RedirectURL)
	}
}

func TestPurchaseValidatesRequest(t *testing.T) {
	gateway, err := New(Config{TerminalID: 12345})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	valid := payment.PurchaseRequest{
		OrderID:     "order-42",
		Amount:      payment.Money{Amount: 125_000, Currency: payment.CurrencyIRR},
		CallbackURL: "https://example.com/callback",
	}

	tests := map[string]payment.PurchaseRequest{
		"missing order ID": func() payment.PurchaseRequest { request := valid; request.OrderID = " "; return request }(),
		"zero amount":      func() payment.PurchaseRequest { request := valid; request.Amount.Amount = 0; return request }(),
		"wrong currency": func() payment.PurchaseRequest {
			request := valid
			request.Amount.Currency = payment.CurrencyIRT
			return request
		}(),
		"unsupported callback scheme": func() payment.PurchaseRequest {
			request := valid
			request.CallbackURL = "ftp://example.com/callback"
			return request
		}(),
		"relative callback": func() payment.PurchaseRequest {
			request := valid
			request.CallbackURL = "/callback"
			return request
		}(),
	}

	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := gateway.Purchase(context.Background(), request)
			if !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestPurchaseReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":-1,"errorCode":"99","errorDesc":"provider failure"}`))
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
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("Purchase() error = %v, want ErrProvider", err)
	}
}

func TestPurchaseRejectsEmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":1,"token":""}`))
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
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("Purchase() error = %v, want ErrProvider", err)
	}
}
