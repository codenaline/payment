package sep

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codenaline/payment"
)

func TestVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			RefNum         string `json:"RefNum"`
			TerminalNumber int64  `json:"TerminalNumber"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.RefNum != "sep-reference" || payload.TerminalNumber != 12345 {
			t.Fatalf("payload = %+v", payload)
		}
		_, _ = w.Write([]byte(`{
			"TransactionDetail": {
				"RRN": "14226761817",
				"RefNum": "sep-reference",
				"MaskedPan": "621986****8080",
				"HashedPan": "card-hash",
				"TerminalNumber": 12345,
				"OrginalAmount": 125000,
				"AffectiveAmount": 125000,
				"StraceDate": "2026-08-28 12:00:00",
				"StraceNo": "100428"
			},
			"ResultCode": 0,
			"ResultDescription": "successful",
			"Success": true
		}`))
	}))
	t.Cleanup(server.Close)

	gateway, err := New(Config{TerminalID: 12345, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	gateway.verifyURL = server.URL

	transaction, err := gateway.Verify(context.Background(), payment.VerifyRequest{
		TransactionID: "sep-reference",
		Amount:        payment.Money{Amount: 125_000, Currency: payment.CurrencyIRR},
	})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if transaction.ID != "sep-reference" || transaction.ReferenceID != "14226761817" || transaction.Status != payment.StatusPaid || transaction.Provider != "sep" || transaction.Amount.Amount != 125_000 || transaction.Amount.Currency != payment.CurrencyIRR {
		t.Fatalf("transaction = %+v", transaction)
	}
}

func TestVerifyValidatesRequest(t *testing.T) {
	gateway, err := New(Config{TerminalID: 12345})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	valid := payment.VerifyRequest{
		TransactionID: "sep-reference",
		Amount:        payment.Money{Amount: 125_000, Currency: payment.CurrencyIRR},
	}
	tests := map[string]payment.VerifyRequest{
		"missing reference": func() payment.VerifyRequest { request := valid; request.TransactionID = " "; return request }(),
		"zero amount":       func() payment.VerifyRequest { request := valid; request.Amount.Amount = 0; return request }(),
		"wrong currency": func() payment.VerifyRequest {
			request := valid
			request.Amount.Currency = payment.CurrencyIRT
			return request
		}(),
	}

	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := gateway.Verify(context.Background(), request)
			if !errors.Is(err, payment.ErrInvalidRequest) {
				t.Fatalf("Verify() error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestVerifyRejectsProviderFailure(t *testing.T) {
	gateway := verificationGateway(t, `{
		"TransactionDetail": {},
		"ResultCode": -999,
		"ResultDescription": "provider failure",
		"Success": false
	}`)

	_, err := gateway.Verify(context.Background(), validVerifyRequest())
	if !errors.Is(err, payment.ErrProvider) {
		t.Fatalf("Verify() error = %v, want ErrProvider", err)
	}
}

func TestVerifyRejectsMismatchedResponse(t *testing.T) {
	tests := map[string]string{
		"reference": `{
			"TransactionDetail": {"RRN":"rrn","RefNum":"other-reference","TerminalNumber":12345,"OrginalAmount":125000},
			"ResultCode":0,"Success":true
		}`,
		"terminal": `{
			"TransactionDetail": {"RRN":"rrn","RefNum":"sep-reference","TerminalNumber":99999,"OrginalAmount":125000},
			"ResultCode":0,"Success":true
		}`,
		"amount": `{
			"TransactionDetail": {"RRN":"rrn","RefNum":"sep-reference","TerminalNumber":12345,"OrginalAmount":124999},
			"ResultCode":0,"Success":true
		}`,
	}

	for name, response := range tests {
		t.Run(name, func(t *testing.T) {
			gateway := verificationGateway(t, response)
			_, err := gateway.Verify(context.Background(), validVerifyRequest())
			if !errors.Is(err, payment.ErrProvider) {
				t.Fatalf("Verify() error = %v, want ErrProvider", err)
			}
		})
	}
}

func verificationGateway(t *testing.T, response string) *Gateway {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	gateway, err := New(Config{TerminalID: 12345, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	gateway.verifyURL = server.URL
	return gateway
}

func validVerifyRequest() payment.VerifyRequest {
	return payment.VerifyRequest{
		TransactionID: "sep-reference",
		Amount:        payment.Money{Amount: 125_000, Currency: payment.CurrencyIRR},
	}
}
