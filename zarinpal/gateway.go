package zarinpal

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codenaline/payment"
)

const (
	defaultAPIBaseURL = "https://api.zarinpal.com/pg/v4/payment"
	defaultPayBaseURL = "https://www.zarinpal.com/pg/StartPay"
	sandboxAPIBaseURL = "https://sandbox.zarinpal.com/pg/v4/payment"
	sandboxPayBaseURL = "https://sandbox.zarinpal.com/pg/StartPay"
)

// Gateway processes payments through Zarinpal.
type Gateway struct {
	merchantID string
	apiBaseURL string
	payBaseURL string
	client     *http.Client
}

// New creates a Zarinpal gateway for merchantID.
func New(merchantID string, options ...Option) (*Gateway, error) {
	if strings.TrimSpace(merchantID) == "" {
		return nil, fmt.Errorf("%w: Zarinpal merchant ID is required", payment.ErrInvalidRequest)
	}

	gateway := &Gateway{
		merchantID: merchantID,
		apiBaseURL: defaultAPIBaseURL,
		payBaseURL: defaultPayBaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("%w: nil Zarinpal option", payment.ErrInvalidRequest)
		}
		if err := option(gateway); err != nil {
			return nil, err
		}
	}

	return gateway, nil
}
