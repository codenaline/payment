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

// New creates a Zarinpal gateway from config.
func New(config Config) (*Gateway, error) {
	if strings.TrimSpace(config.MerchantID) == "" {
		return nil, fmt.Errorf("%w: Zarinpal merchant ID is required", payment.ErrInvalidRequest)
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	gateway := &Gateway{
		merchantID: config.MerchantID,
		apiBaseURL: defaultAPIBaseURL,
		payBaseURL: defaultPayBaseURL,
		client:     client,
	}
	if config.Sandbox {
		gateway.apiBaseURL = sandboxAPIBaseURL
		gateway.payBaseURL = sandboxPayBaseURL
	}

	return gateway, nil
}
