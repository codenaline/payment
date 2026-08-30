package aqayepardakht

import "net/http"

// Config configures an Aghaye Pardakht gateway.
type Config struct {
	Pin        string
	Sandbox    bool
	HTTPClient *http.Client
}
