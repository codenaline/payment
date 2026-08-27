package nextpay

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

// Refund refunds a verified NextPay transaction.
func (g *Gateway) Refund(ctx context.Context, request payment.RefundRequest) (payment.RefundResponse, error) {
	if err := validateVerify(payment.VerifyRequest{TransactionID: request.TransactionID, Amount: request.Amount}); err != nil {
		return payment.RefundResponse{}, err
	}

	form := url.Values{
		"api_key":        {g.apiKey},
		"trans_id":       {request.TransactionID},
		"amount":         {strconv.FormatInt(request.Amount.Amount, 10)},
		"currency":       {strings.ToUpper(request.Amount.Currency)},
		"refund_request": {"yes_money_back"},
	}
	var response verifyResponse
	if err := g.post(ctx, "/verify", form, &response); err != nil {
		return payment.RefundResponse{}, newError("refund", 0, "request failed", err)
	}
	if response.Code != int(CodeRefunded) {
		return payment.RefundResponse{}, newError("refund", response.Code, message(Code(response.Code)), nil)
	}

	return payment.RefundResponse{
		ID:            request.TransactionID,
		TransactionID: request.TransactionID,
		Amount:        request.Amount,
	}, nil
}

var _ payment.Refunder = (*Gateway)(nil)
