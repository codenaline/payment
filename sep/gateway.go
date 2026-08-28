package sep

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codenaline/payment"
)

const (
	defaultTokenURL  = "https://sep.shaparak.ir/onlinepg/onlinepg"
	defaultPayURL    = "https://sep.shaparak.ir/OnlinePG/SendToken"
	defaultVerifyURL = "https://sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction"
)

// Gateway processes payments through Saman Electronic Payment (SEP).
type Gateway struct {
	terminalID int64
	tokenURL   string
	payURL     string
	verifyURL  string
	client     *http.Client
}

// New creates an SEP gateway from config.
func New(config Config) (*Gateway, error) {
	if config.TerminalID <= 0 {
		return nil, fmt.Errorf("%w: SEP terminal ID must be positive", payment.ErrInvalidRequest)
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &Gateway{
		terminalID: config.TerminalID,
		tokenURL:   defaultTokenURL,
		payURL:     defaultPayURL,
		verifyURL:  defaultVerifyURL,
		client:     client,
	}, nil
}
