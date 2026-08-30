package aqayepardakht

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codenaline/payment"
)

const (
	defaultCreateURL     = "https://panel.aqayepardakht.ir/api/v2/create"
	defaultVerifyURL     = "https://panel.aqayepardakht.ir/api/v2/verify"
	defaultPayURL        = "https://panel.aqayepardakht.ir/startpay"
	defaultSandboxPayURL = "https://panel.aqayepardakht.ir/startpay/sandbox"
)

// Gateway processes payments through Aghaye Pardakht.
type Gateway struct {
	pin       string
	sandbox   bool
	createURL string
	verifyURL string
	payURL    string
	client    *http.Client
}

// New creates an Aghaye Pardakht gateway from config.
func New(config Config) (*Gateway, error) {
	if strings.TrimSpace(config.Pin) == "" {
		return nil, fmt.Errorf("%w: Aghaye Pardakht PIN is required", payment.ErrInvalidRequest)
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	payURL := defaultPayURL
	if config.Sandbox {
		payURL = defaultSandboxPayURL
	}

	return &Gateway{
		pin:       config.Pin,
		sandbox:   config.Sandbox,
		createURL: defaultCreateURL,
		verifyURL: defaultVerifyURL,
		payURL:    payURL,
		client:    client,
	}, nil
}
