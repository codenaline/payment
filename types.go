package payment

import "time"

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

// Money represents an amount in the currency's smallest unit.
// For example, USD 10.50 is represented as Amount 1050 and Currency "USD".
type Money struct {
	Amount   int64
	Currency string
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

// VerifyRequest identifies a transaction and includes data returned by the
// payment provider.
type VerifyRequest struct {
	TransactionID string
	Amount        Money
	Data          map[string]string
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

type InquiryRequest struct {
	TransactionID string
}

// Transaction represents a payment-provider transaction.
type Transaction struct {
	ID          string
	Amount      Money
	Status      Status
	ReferenceID string
	Provider    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
