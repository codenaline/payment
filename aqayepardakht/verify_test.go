package aqayepardakht

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestVerifyReturnsPaidTransaction(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if r.Form.Get("pin") != "merchant-pin" || r.Form.Get("amount") != "12500" || r.Form.Get("transid") != "transaction-token" {
			t.Errorf("request form = %v", r.Form)
		}
		_, _ = w.Write([]byte(`{"status":"success","code":1}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "merchant-pin", HTTPClient: server.Client()})
	gateway.verifyURL = server.URL
	transaction, err := gateway.Verify(t.Context(), validVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if transaction.ID != "transaction-token" || transaction.Provider != providerName || transaction.Status != payment.StatusPaid {
		t.Errorf("Verify() transaction = %+v", transaction)
	}
	if transaction.ReferenceID != "" {
		t.Errorf("Verify() reference ID = %q, want empty", transaction.ReferenceID)
	}
	if transaction.Amount != (payment.Money{Amount: 125000, Currency: payment.CurrencyIRR}) {
		t.Errorf("Verify() amount = %+v", transaction.Amount)
	}
}

func TestVerifyTreatsAlreadyVerifiedAsPaid(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","code":"2"}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "pin", HTTPClient: server.Client()})
	gateway.verifyURL = server.URL
	transaction, err := gateway.Verify(t.Context(), validVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if transaction.Status != payment.StatusPaid {
		t.Errorf("Verify() status = %q, want paid", transaction.Status)
	}
}

func TestVerifyClassifiesProviderErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want error
		code Code
	}{
		{name: "payment failed", body: `{"status":"error","code":0}`, want: payment.ErrDeclined, code: CodePaymentFailed},
		{name: "transaction missing", body: `{"status":"error","code":-8}`, want: payment.ErrTransactionNotFound, code: CodeTransactionNotFound},
		{name: "amount mismatch", body: `{"status":"error","code":-10}`, want: payment.ErrInvalidRequest, code: CodeAmountMismatch},
		{name: "gateway inactive", body: `{"status":"error","code":-11}`, want: payment.ErrProvider, code: CodeGatewayInactive},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			gateway, _ := New(Config{Pin: "pin", HTTPClient: server.Client()})
			gateway.verifyURL = server.URL
			_, err := gateway.Verify(t.Context(), validVerifyRequest())
			if !errors.Is(err, test.want) {
				t.Fatalf("Verify() error = %v, want %v", err, test.want)
			}
			var providerError *Error
			if !errors.As(err, &providerError) || providerError.Code != test.code {
				t.Fatalf("errors.As() = %#v, want code %d", providerError, test.code)
			}
		})
	}
}

func TestVerifyRejectsContradictorySuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","code":-8}`))
	}))
	defer server.Close()

	gateway, _ := New(Config{Pin: "pin", HTTPClient: server.Client()})
	gateway.verifyURL = server.URL
	_, err := gateway.Verify(t.Context(), validVerifyRequest())
	if !errors.Is(err, payment.ErrTransactionNotFound) {
		t.Fatalf("Verify() error = %v, want ErrTransactionNotFound", err)
	}
}

func TestVerifyValidatesRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*payment.VerifyRequest)
	}{
		{name: "missing transaction ID", mutate: func(r *payment.VerifyRequest) { r.TransactionID = " " }},
		{name: "invalid amount", mutate: func(r *payment.VerifyRequest) { r.Amount.Amount = 0 }},
		{name: "unsupported currency", mutate: func(r *payment.VerifyRequest) { r.Amount.Currency = "USD" }},
	}

	gateway, _ := New(Config{Pin: "pin"})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validVerifyRequest()
			test.mutate(&request)
			_, err := gateway.Verify(t.Context(), request)
			if !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("Verify() error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func validVerifyRequest() payment.VerifyRequest {
	return payment.VerifyRequest{
		TransactionID: "transaction-token",
		Amount:        payment.Money{Amount: 125000, Currency: payment.CurrencyIRR},
	}
}
