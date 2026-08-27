package payment

// Status describes the current state of a transaction.
type Status string

const (
	// StatusPending indicates that a transaction has not completed yet.
	StatusPending Status = "pending"
	// StatusPaid indicates that a transaction completed successfully.
	StatusPaid Status = "paid"
	// StatusFailed indicates that a transaction failed.
	StatusFailed Status = "failed"
	// StatusRefunded indicates that a transaction was refunded.
	StatusRefunded Status = "refunded"
	// StatusCanceled indicates that a transaction was canceled.
	StatusCanceled Status = "canceled"
)

// Currency is a payment currency code.
// It remains open-ended so applications and third-party gateways can use
// currencies that are not predefined by this package.
type Currency string

const (
	CurrencyIRR Currency = "IRR"
	CurrencyIRT Currency = "IRT"
)

// Money represents an amount in the currency's smallest unit.
type Money struct {
	Amount   int64
	Currency Currency
}

// PurchaseRequest contains the information needed to initiate a payment.
type PurchaseRequest struct {
	OrderID     string
	Amount      Money
	CallbackURL string
	Description string
	Metadata    map[string]string
}

// PurchaseResponse contains the newly created transaction and the URL where
// the payer should be redirected.
type PurchaseResponse struct {
	Transaction Transaction
	RedirectURL string
}

// VerifyRequest contains the information needed to verify a transaction.
type VerifyRequest struct {
	TransactionID string
	Amount        Money
}

// RefundRequest contains the information needed to issue a full or partial
// refund.
type RefundRequest struct {
	TransactionID string
	Amount        Money
	Reason        string
}

// RefundResponse identifies a refund accepted by the payment provider.
type RefundResponse struct {
	ID            string
	TransactionID string
	Amount        Money
}

// Transaction represents a payment-provider transaction.
type Transaction struct {
	ID          string
	Amount      Money
	Status      Status
	ReferenceID string
	Provider    string
}
