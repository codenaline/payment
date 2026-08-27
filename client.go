package payment

import "context"

// Client provides payment operations through one immutable gateway. Create a
// separate Client for each configured gateway.
type Client struct {
	gateway Gateway
}

// NewClient creates a client backed by gateway. It panics if gateway is nil.
func NewClient(gateway Gateway) *Client {
	if gateway == nil {
		panic("payment: nil gateway")
	}
	return &Client{gateway: gateway}
}

// Purchase initiates a payment through the client's gateway.
func (c *Client) Purchase(ctx context.Context, request PurchaseRequest) (PurchaseResponse, error) {
	return c.gateway.Purchase(ctx, request)
}

// Verify verifies a payment through the client's gateway.
func (c *Client) Verify(ctx context.Context, request VerifyRequest) (Transaction, error) {
	return c.gateway.Verify(ctx, request)
}

// Refund refunds a payment if the client's gateway supports refunds.
func (c *Client) Refund(ctx context.Context, request RefundRequest) (RefundResponse, error) {
	refunder, ok := c.gateway.(Refunder)
	if !ok {
		return RefundResponse{}, ErrUnsupported
	}
	return refunder.Refund(ctx, request)
}

// Inquiry retrieves a transaction if the client's gateway supports
// transaction lookup.
func (c *Client) Inquiry(ctx context.Context, req InquiryRequest) (Transaction, error) {
	inquirer, ok := c.gateway.(Inquirer)
	if !ok {
		return Transaction{}, ErrUnsupported
	}
	return inquirer.Inquiry(ctx, req)
}
