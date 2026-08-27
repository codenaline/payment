package zarinpal

import (
	"fmt"
	"net/http"

	"github.com/codenaline/payment"
)

// Option configures a Gateway.
type Option func(*Gateway) error

// WithHTTPClient configures the client used for Zarinpal API requests.
func WithHTTPClient(client *http.Client) Option {
	return func(gateway *Gateway) error {
		if client == nil {
			return fmt.Errorf("%w: nil Zarinpal HTTP client", payment.ErrInvalidRequest)
		}
		gateway.client = client
		return nil
	}
}

// WithSandbox configures the gateway to use Zarinpal's sandbox.
func WithSandbox() Option {
	return func(gateway *Gateway) error {
		gateway.apiBaseURL = sandboxAPIBaseURL
		gateway.payBaseURL = sandboxPayBaseURL
		return nil
	}
}
