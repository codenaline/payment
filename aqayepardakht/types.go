package aqayepardakht

type apiResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	TransID string `json:"transid,omitempty"`
}
