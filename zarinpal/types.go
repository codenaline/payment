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

func decodeAPIError(raw json.RawMessage) apiError {
	var result apiError
	_ = json.Unmarshal(raw, &result)
	return result
}
