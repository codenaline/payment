package nextpay

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestPurchase(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/token" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if r.Form.Get("auto_verify") != "" {
			t.Errorf("request unexpectedly enables auto_verify")
		}
		if r.Form.Get("api_key") != "key" || r.Form.Get("order_id") != "order-1" || r.Form.Get("amount") != "12500" || r.Form.Get("currency") != "IRR" || r.Form.Get("customer_phone") != "09121234567" {
			t.Errorf("request form = %v", r.Form)
		}
		_, _ = w.Write([]byte(`{"code":-1,"trans_id":"transaction-token"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL
	gateway.payBaseURL = "https://pay.example"
	result, err := gateway.Purchase(t.Context(), validPurchaseRequest())
	if err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}
	if result.Transaction.ID != "transaction-token" || result.Transaction.Provider != providerName || result.Transaction.Status != payment.StatusPending {
		t.Errorf("Purchase() transaction = %+v", result.Transaction)
	}
	if result.RedirectURL != "https://pay.example/transaction-token" {
		t.Errorf("Purchase() redirect URL = %q", result.RedirectURL)
	}
}

func TestPurchaseReturnsProviderError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":-24,"trans_id":""}`))
	}))
	defer server.Close()
	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL

	_, err := gateway.Purchase(t.Context(), validPurchaseRequest())
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
	}
	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Code != CodeInvalidAmount {
		t.Fatalf("errors.As() = %#v, want CodeInvalidAmount", providerError)
	}
}

func TestPurchaseValidatesRequest(t *testing.T) {
	t.Parallel()

	gateway, _ := New(Config{APIKey: "key"})
	request := validPurchaseRequest()
	request.OrderID = ""
	_, err := gateway.Purchase(t.Context(), request)
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
	}
}

func validPurchaseRequest() payment.PurchaseRequest {
	return payment.PurchaseRequest{
		OrderID:     "order-1",
		Amount:      payment.Money{Amount: 12500, Currency: payment.CurrencyIRR},
		CallbackURL: "https://merchant.example/callback",
		Description: "order",
		Metadata: map[string]string{
			"customer_phone": "09121234567",
			"auto_verify":    "yes",
		},
	}
}
