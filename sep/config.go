package sep

import "net/http"

// Config configures an SEP gateway.
type Config struct {
	TerminalID int64
	HTTPClient *http.Client
}
