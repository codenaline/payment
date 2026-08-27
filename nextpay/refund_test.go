package nextpay

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestRefund(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if r.URL.Path != "/verify" || r.Form.Get("refund_request") != "yes_money_back" || r.Form.Get("trans_id") != "transaction-token" {
			t.Errorf("request = %s %v", r.URL.Path, r.Form)
		}
		_, _ = w.Write([]byte(`{"code":-90,"amount":12500,"order_id":"order-1"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL
	result, err := gateway.Refund(t.Context(), validRefundRequest())
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if result.ID != "" || result.TransactionID != "transaction-token" || result.Amount.Amount != 12500 {
		t.Errorf("Refund() result = %+v", result)
	}
}

func TestRefundReturnsProviderError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":-91}`))
	}))
	defer server.Close()
	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL

	_, err := gateway.Refund(t.Context(), validRefundRequest())
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("Refund() error = %v, want ErrProvider", err)
	}
	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Code != CodeRefundFailed {
		t.Fatalf("errors.As() = %#v, want CodeRefundFailed", providerError)
	}
}

func validRefundRequest() payment.RefundRequest {
	return payment.RefundRequest{
		TransactionID: "transaction-token",
		Amount:        payment.Money{Amount: 12500, Currency: payment.CurrencyIRR},
	}
}
