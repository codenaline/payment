package payment

// Money represents an amount in the currency's smallest unit.
// For example, USD 10.50 is represented as Amount 1050 and Currency "USD".
type Money struct {
	Amount   int64
	Currency string
}

// PurchaseRequest contains the information needed to initiate a payment.
type PurchaseRequest struct {
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

// Transaction represents a payment-provider transaction.
type Transaction struct {
	ID     string
	Amount Money
	Status string
}
