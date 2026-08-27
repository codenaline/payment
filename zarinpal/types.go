package zarinpal

import "encoding/json"

type requestPayload struct {
	MerchantID  string            `json:"merchant_id"`
	Amount      int64             `json:"amount"`
	CallbackURL string            `json:"callback_url"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type requestResponse struct {
	Data struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Authority string `json:"authority"`
	} `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type verifyPayload struct {
	MerchantID string `json:"merchant_id"`
	Amount     int64  `json:"amount"`
	Authority  string `json:"authority"`
}

type verifyResponse struct {
	Data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		RefID   int64  `json:"ref_id"`
	} `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

func decodeAPIError(raw json.RawMessage) apiError {
	var result apiError
	_ = json.Unmarshal(raw, &result)
	return result
}
