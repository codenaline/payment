package payment

import "context"

// Gateway defines the operations required from every payment provider.
type Gateway interface {
	Purchaser
	Verifier
}

// Purchaser initiates payments.
type Purchaser interface {
	Purchase(context.Context, PurchaseRequest) (PurchaseResponse, error)
}

// Verifier verifies payments.
type Verifier interface {
	Verify(context.Context, VerifyRequest) (Transaction, error)
}

// Refunder refunds payments when supported by a provider.
type Refunder interface {
	Refund(context.Context, RefundRequest) (RefundResponse, error)
}

// Inquirer retrieves transactions when supported by a provider.
type Inquirer interface {
	Inquiry(context.Context, InquiryRequest) (Transaction, error)
}
