package nextpay

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codenaline/payment"
)

const (
	defaultAPIBaseURL = "https://nextpay.org/nx/gateway"
	defaultPayBaseURL = "https://nextpay.org/nx/gateway/payment"
)

// Gateway processes payments through NextPay.
type Gateway struct {
	apiKey     string
	apiBaseURL string
	payBaseURL string
	client     *http.Client
}

// New creates a NextPay gateway from config.
func New(config Config) (*Gateway, error) {
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, fmt.Errorf("%w: NextPay API key is required", payment.ErrInvalidRequest)
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &Gateway{
		apiKey:     config.APIKey,
		apiBaseURL: defaultAPIBaseURL,
		payBaseURL: defaultPayBaseURL,
		client:     client,
	}, nil
}
