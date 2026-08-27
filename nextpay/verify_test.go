package nextpay

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify" || r.Method != http.MethodPost {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if r.Form.Get("trans_id") != "transaction-token" || r.Form.Get("amount") != "12500" {
			t.Errorf("request form = %v", r.Form)
		}
		_, _ = w.Write([]byte(`{"code":0,"amount":12500,"order_id":"order-1","Shaparak_Ref_Id":"reference-1"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL
	transaction, err := gateway.Verify(t.Context(), validVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if transaction.ID != "transaction-token" || transaction.ReferenceID != "reference-1" || transaction.Status != payment.StatusPaid {
		t.Errorf("Verify() transaction = %+v", transaction)
	}
}

func TestVerifyReturnsProviderError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":-37}`))
	}))
	defer server.Close()
	gateway, _ := New(Config{APIKey: "key", HTTPClient: server.Client()})
	gateway.apiBaseURL = server.URL

	_, err := gateway.Verify(t.Context(), validVerifyRequest())
	if !errors.Is(err, payment.ErrTransactionNotFound) {
		t.Fatalf("Verify() error = %v, want ErrTransactionNotFound", err)
	}
	var providerError *Error
	if !errors.As(err, &providerError) || providerError.Code != CodeTransactionNotFound {
		t.Fatalf("errors.As() = %#v, want CodeTransactionNotFound", providerError)
	}
}

func TestVerifyValidatesRequest(t *testing.T) {
	t.Parallel()

	gateway, _ := New(Config{APIKey: "key"})
	request := validVerifyRequest()
	request.TransactionID = ""
	_, err := gateway.Verify(t.Context(), request)
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Verify() error = %v, want ErrInvalidRequest", err)
	}
}

func validVerifyRequest() payment.VerifyRequest {
	return payment.VerifyRequest{
		TransactionID: "transaction-token",
		Amount:        payment.Money{Amount: 12500, Currency: payment.CurrencyIRR},
	}
}
