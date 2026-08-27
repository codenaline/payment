package payment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codenaline/payment"
)

type gatewayStub struct {
	id string
}

func (g *gatewayStub) Purchase(context.Context, payment.PurchaseRequest) (payment.PurchaseResponse, error) {
	return payment.PurchaseResponse{Transaction: payment.Transaction{Provider: g.id}}, nil
}

func (g *gatewayStub) Verify(context.Context, payment.VerifyRequest) (payment.Transaction, error) {
	return payment.Transaction{Provider: g.id}, nil
}

type refundingGateway struct {
	*gatewayStub
}

func (*refundingGateway) Refund(context.Context, payment.RefundRequest) (payment.RefundResponse, error) {
	return payment.RefundResponse{ID: "refund"}, nil
}

func TestClientsUseIndependentGateways(t *testing.T) {
	t.Parallel()

	first := payment.NewClient(&gatewayStub{id: "first"})
	second := payment.NewClient(&gatewayStub{id: "second"})

	firstResult, err := first.Purchase(t.Context(), payment.PurchaseRequest{})
	if err != nil {
		t.Fatalf("first Purchase() error = %v", err)
	}
	secondResult, err := second.Purchase(t.Context(), payment.PurchaseRequest{})
	if err != nil {
		t.Fatalf("second Purchase() error = %v", err)
	}

	if firstResult.Transaction.Provider != "first" {
		t.Errorf("first provider = %q", firstResult.Transaction.Provider)
	}
	if secondResult.Transaction.Provider != "second" {
		t.Errorf("second provider = %q", secondResult.Transaction.Provider)
	}
}

func TestClientReturnsUnsupportedForMissingCapability(t *testing.T) {
	t.Parallel()

	client := payment.NewClient(&gatewayStub{})
	_, err := client.Refund(t.Context(), payment.RefundRequest{})
	if !errors.Is(err, payment.ErrUnsupported) {
		t.Fatalf("Refund() error = %v, want ErrUnsupported", err)
	}
}

func TestClientDelegatesOptionalCapability(t *testing.T) {
	t.Parallel()

	client := payment.NewClient(&refundingGateway{gatewayStub: &gatewayStub{}})
	result, err := client.Refund(t.Context(), payment.RefundRequest{})
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if result.ID != "refund" {
		t.Errorf("Refund() ID = %q", result.ID)
	}
}

func TestNewClientPanicsForNilGateway(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("NewClient() did not panic")
		}
	}()
	payment.NewClient(nil)
}
