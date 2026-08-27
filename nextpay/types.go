package nextpay

type tokenResponse struct {
	Code    int    `json:"code"`
	TransID string `json:"trans_id"`
}

type verifyResponse struct {
	Code          int    `json:"code"`
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	ShaparakRefID string `json:"Shaparak_Ref_Id"`
}
