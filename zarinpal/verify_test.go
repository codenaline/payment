package zarinpal

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify.json" || r.Method != http.MethodPost {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		var body verifyPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body.Authority != "A0001" || body.Amount != 12500 || body.MerchantID != "merchant" {
			t.Errorf("request body = %+v", body)
		}
		_, _ = w.Write([]byte(`{"data":{"code":100,"ref_id":987654},"errors":[]}`))
	}))
	defer server.Close()

	gateway, _ := New("merchant", WithHTTPClient(server.Client()))
	gateway.apiBaseURL = server.URL
	transaction, err := gateway.Verify(t.Context(), payment.VerifyRequest{
		TransactionID: "A0001",
		Amount:        payment.Money{Amount: 12500, Currency: "IRR"},
	})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if transaction.ID != "A0001" || transaction.ReferenceID != "987654" || transaction.Status != payment.StatusPaid {
		t.Errorf("Verify() transaction = %+v", transaction)
	}
}

func TestVerifyAcceptsAlreadyVerified(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"code":101,"ref_id":42},"errors":[]}`))
	}))
	defer server.Close()
	gateway, _ := New("merchant", WithHTTPClient(server.Client()))
	gateway.apiBaseURL = server.URL

	transaction, err := gateway.Verify(t.Context(), validVerifyRequest())
	if err != nil || transaction.Status != payment.StatusPaid {
		t.Fatalf("Verify() = %+v, %v", transaction, err)
	}
}

func TestVerifyReturnsGatewayError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{},"errors":{"code":-54,"message":"request archived"}}`))
	}))
	defer server.Close()
	gateway, _ := New("merchant", WithHTTPClient(server.Client()))
	gateway.apiBaseURL = server.URL

	_, err := gateway.Verify(t.Context(), validVerifyRequest())
	if !errors.Is(err, payment.ErrTransactionNotFound) {
		t.Fatalf("Verify() error = %v, want ErrTransactionNotFound", err)
	}
	code, ok := CodeOf(err)
	if !ok || code != CodeRequestArchived {
		t.Fatalf("CodeOf() = %q, %t; want %q, true", code, ok, CodeRequestArchived)
	}
}

func TestVerifyValidatesRequest(t *testing.T) {
	t.Parallel()

	gateway, _ := New("merchant")
	request := validVerifyRequest()
	request.TransactionID = ""
	_, err := gateway.Verify(t.Context(), request)
	if !errors.Is(err, payment.ErrInvalidRequest) {
		t.Fatalf("Verify() error = %v, want ErrInvalidRequest", err)
	}
}

func validVerifyRequest() payment.VerifyRequest {
	return payment.VerifyRequest{
		TransactionID: "A0001",
		Amount:        payment.Money{Amount: 1000, Currency: "IRR"},
	}
}
