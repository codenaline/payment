package aqayepardakht

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

// Verify verifies an Aghaye Pardakht transaction.
func (g *Gateway) Verify(ctx context.Context, request payment.VerifyRequest) (payment.Transaction, error) {
	amount, err := validateVerify(request)
	if err != nil {
		return payment.Transaction{}, err
	}

	form := url.Values{
		"pin":     {g.pin},
		"amount":  {strconv.FormatInt(amount, 10)},
		"transid": {request.TransactionID},
	}
	var response apiResponse
	if err := g.postForm(ctx, g.verifyURL, form, &response); err != nil {
		return payment.Transaction{}, newError("verification", 0, "request failed", err)
	}
	if !strings.EqualFold(response.Status, "success") || (response.Code != CodeSuccess && response.Code != CodeAlreadyVerified) {
		return payment.Transaction{}, newError("verification", response.Code, message(response.Code), nil)
	}

	return payment.Transaction{
		ID:       request.TransactionID,
		Amount:   request.Amount,
		Status:   payment.StatusPaid,
		Provider: providerName,
	}, nil
}

func validateVerify(request payment.VerifyRequest) (int64, error) {
	if strings.TrimSpace(request.TransactionID) == "" {
		return 0, fmt.Errorf("%w: transaction ID is required", payment.ErrInvalidRequest)
	}
	return normalizeAmount(request.Amount)
}

var _ payment.Gateway = (*Gateway)(nil)
