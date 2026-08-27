package nextpay

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

// Verify verifies and settles a NextPay transaction.
func (g *Gateway) Verify(ctx context.Context, request payment.VerifyRequest) (payment.Transaction, error) {
	if err := validateVerify(request); err != nil {
		return payment.Transaction{}, err
	}

	form := url.Values{
		"api_key":  {g.apiKey},
		"trans_id": {request.TransactionID},
		"amount":   {strconv.FormatInt(request.Amount.Amount, 10)},
		"currency": {strings.ToUpper(request.Amount.Currency)},
	}
	var response verifyResponse
	if err := g.post(ctx, "/verify", form, &response); err != nil {
		return payment.Transaction{}, newError("verification", 0, "request failed", err)
	}
	if response.Code != 0 {
		return payment.Transaction{}, newError("verification", response.Code, message(Code(response.Code)), nil)
	}
	if response.Amount != 0 && response.Amount != request.Amount.Amount {
		return payment.Transaction{}, newError("verification", 0, "provider returned a different amount", payment.ErrProvider)
	}

	return payment.Transaction{
		ID:          request.TransactionID,
		Amount:      request.Amount,
		Status:      payment.StatusPaid,
		ReferenceID: response.ShaparakRefID,
		Provider:    providerName,
	}, nil
}

func validateVerify(request payment.VerifyRequest) error {
	if strings.TrimSpace(request.TransactionID) == "" {
		return fmt.Errorf("%w: transaction ID is required", payment.ErrInvalidRequest)
	}
	if request.Amount.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", payment.ErrInvalidRequest)
	}
	currency := strings.ToUpper(request.Amount.Currency)
	if currency != "IRR" && currency != "IRT" {
		return fmt.Errorf("%w: NextPay requires IRR or IRT", payment.ErrInvalidRequest)
	}
	return nil
}

var _ payment.Gateway = (*Gateway)(nil)
