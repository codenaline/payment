package payment

import "context"

// Gateway processes payments through a payment provider.
type Gateway interface {
	Purchase(context.Context, PurchaseRequest) (PurchaseResponse, error)
	Verify(context.Context, VerifyRequest) (Transaction, error)
	Refund(context.Context, RefundRequest) (RefundResponse, error)
	GetTransaction(context.Context, string) (Transaction, error)
}
