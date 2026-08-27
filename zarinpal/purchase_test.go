package zarinpal

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestPurchase(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/request.json" || r.Method != http.MethodPost {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		var body requestPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body.MerchantID != "merchant" || body.Amount != 12500 {
			t.Errorf("request body = %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"code":100,"authority":"A0001"},"errors":[]}`))
	}))
	defer server.Close()

	gateway, _ := New("merchant", WithHTTPClient(server.Client()))
	gateway.apiBaseURL = server.URL
	gateway.payBaseURL = "https://pay.example"

	result, err := gateway.Purchase(t.Context(), payment.PurchaseRequest{
		Amount:      payment.Money{Amount: 12500, Currency: "IRR"},
		CallbackURL: "https://merchant.example/callback",
		Description: "order 42",
	})
	if err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}
	if result.Transaction.ID != "A0001" || result.Transaction.Status != payment.StatusPending {
		t.Errorf("Purchase() transaction = %+v", result.Transaction)
	}
	if result.RedirectURL != "https://pay.example/A0001" {
		t.Errorf("Purchase() redirect URL = %q", result.RedirectURL)
	}
}

func TestPurchaseReturnsGatewayError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{},"errors":{"code":-9,"message":"validation failed"}}`))
	}))
	defer server.Close()
	gateway, _ := New("merchant", WithHTTPClient(server.Client()))
	gateway.apiBaseURL = server.URL

	_, err := gateway.Purchase(t.Context(), validPurchaseRequest())
	if !errors.Is(err, payment.ErrPurchaseFailed) {
		t.Fatalf("Purchase() error = %v, want ErrPurchaseFailed", err)
	}
	var providerErr *payment.Error
	if !errors.As(err, &providerErr) || providerErr.Code != "-9" {
		t.Fatalf("Purchase() error = %#v, want Zarinpal code -9", err)
	}
}

func TestPurchaseValidatesRequest(t *testing.T) {
	t.Parallel()

	gateway, _ := New("merchant")
	request := validPurchaseRequest()
	request.Amount.Amount = 0
	_, err := gateway.Purchase(t.Context(), request)
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Purchase() error = %v, want ErrInvalidRequest", err)
	}
}

func validPurchaseRequest() payment.PurchaseRequest {
	return payment.PurchaseRequest{
		Amount:      payment.Money{Amount: 1000, Currency: "IRR"},
		CallbackURL: "https://merchant.example/callback",
		Description: "order",
	}
}
