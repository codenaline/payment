package zarinpal

import "net/http"

// Config configures a Zarinpal gateway.
type Config struct {
	MerchantID string
	Sandbox    bool
	HTTPClient *http.Client
}
