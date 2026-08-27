package nextpay

import "net/http"

// Config configures a NextPay gateway.
type Config struct {
	APIKey     string
	HTTPClient *http.Client
}
