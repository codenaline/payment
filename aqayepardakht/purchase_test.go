package aqayepardakht

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestPurchaseCreatesTransactionInTomanFromRial(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		want := map[string]string{
			"pin":         "merchant-pin",
			"amount":      "12500",
			"callback":    "https://merchant.example/callback",
			"invoice_id":  "order-1",
			"description": "Order #1",
			"card_number": "6037991234567890",
			"mobile":      "09121234567",
			"email":       "payer@example.com",
		}
		for key, value := range want {
			if got := r.Form.Get(key); got != value {
				t.Errorf("form[%q] = %q, want %q", key, got, value)
			}
		}
		if got := r.Form.Get("private_note"); got != "" {
			t.Errorf("private metadata was forwarded: %q", got)
		}
		_, _ = w.Write([]byte(`{"status":"success","transid":"transaction-token"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "merchant-pin", HTTPClient: server.Client()})
	gateway.createURL = server.URL
	gateway.payURL = "https://pay.example/startpay"
	result, err := gateway.Purchase(t.Context(), validPurchaseRequest())
	if err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}
	if result.Transaction.ID != "transaction-token" || result.Transaction.Provider != providerName || result.Transaction.Status != payment.StatusPending {
		t.Errorf("Purchase() transaction = %+v", result.Transaction)
	}
	if result.Transaction.Amount != (payment.Money{Amount: 125000, Currency: payment.CurrencyIRR}) {
		t.Errorf("Purchase() amount = %+v", result.Transaction.Amount)
	}
	if result.RedirectURL != "https://pay.example/startpay/transaction-token" {
		t.Errorf("Purchase() redirect URL = %q", result.RedirectURL)
	}
}

func TestPurchaseUsesTomanAndSandboxRedirect(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("amount"); got != "12500" {
			t.Errorf("amount = %q, want 12500", got)
		}
		_, _ = w.Write([]byte(`{"status":"success","transid":"token/with space"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "pin", Sandbox: true, HTTPClient: server.Client()})
	gateway.createURL = server.URL
	request := validPurchaseRequest()
	request.Amount = payment.Money{Amount: 12500, Currency: payment.CurrencyIRT}
	result, err := gateway.Purchase(t.Context(), request)
	if err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}
	if result.RedirectURL != defaultSandboxPayURL+"/token%2Fwith%20space" {
		t.Errorf("Purchase() redirect URL = %q", result.RedirectURL)
	}
}

func TestPurchaseReturnsTypedProviderError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","code":"-5"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "pin", HTTPClient: server.Client()})
	gateway.createURL = server.URL
	_, err := gateway.Purchase(t.Context(), validPurchaseRequest())
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
	}
	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Code != CodeAmountOutOfRange {
		t.Fatalf("errors.As() = %#v, want CodeAmountOutOfRange", providerError)
	}
}

func TestPurchaseValidatesRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*payment.PurchaseRequest)
	}{
		{name: "missing order ID", mutate: func(r *payment.PurchaseRequest) { r.OrderID = " " }},
		{name: "invalid callback", mutate: func(r *payment.PurchaseRequest) { r.CallbackURL = "ftp://merchant.example/callback" }},
		{name: "amount below minimum", mutate: func(r *payment.PurchaseRequest) { r.Amount.Amount = 990 }},
		{name: "amount above maximum", mutate: func(r *payment.PurchaseRequest) { r.Amount.Amount = 500000010 }},
		{name: "fractional toman", mutate: func(r *payment.PurchaseRequest) { r.Amount.Amount = 1001 }},
		{name: "unsupported currency", mutate: func(r *payment.PurchaseRequest) { r.Amount.Currency = "USD" }},
	}

	gateway, _ := New(Config{Pin: "pin"})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validPurchaseRequest()
			test.mutate(&request)
			_, err := gateway.Purchase(t.Context(), request)
			if !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func validPurchaseRequest() payment.PurchaseRequest {
	return payment.PurchaseRequest{
		OrderID:     "order-1",
		Amount:      payment.Money{Amount: 125000, Currency: payment.CurrencyIRR},
		CallbackURL: "https://merchant.example/callback",
		Description: "Order #1",
		Metadata: map[string]string{
			"card_number":  "6037991234567890",
			"mobile":       "09121234567",
			"email":        "payer@example.com",
			"private_note": "do not send",
		},
	}
}
