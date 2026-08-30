package aqayepardakht

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

const (
	providerName  = "aqayepardakht"
	minimumAmount = int64(100)
	maximumAmount = int64(50_000_000)
)

// Purchase creates an Aghaye Pardakht transaction.
func (g *Gateway) Purchase(ctx context.Context, request payment.PurchaseRequest) (payment.PurchaseResponse, error) {
	amount, err := validatePurchase(request)
	if err != nil {
		return payment.PurchaseResponse{}, err
	}

	form := url.Values{
		"pin":        {g.pin},
		"amount":     {strconv.FormatInt(amount, 10)},
		"callback":   {request.CallbackURL},
		"invoice_id": {request.OrderID},
	}
	if request.Description != "" {
		form.Set("description", request.Description)
	}
	for _, field := range []string{"card_number", "mobile", "email"} {
		if value := request.Metadata[field]; value != "" {
			form.Set(field, value)
		}
	}

	var response apiResponse
	if err := g.postForm(ctx, g.createURL, form, &response); err != nil {
		return payment.PurchaseResponse{}, newError("purchase", 0, "request failed", err)
	}
	if !strings.EqualFold(response.Status, "success") || strings.TrimSpace(response.TransID) == "" {
		return payment.PurchaseResponse{}, newError("purchase", response.Code, message(response.Code), nil)
	}

	return payment.PurchaseResponse{
		Transaction: payment.Transaction{
			ID:       response.TransID,
			Amount:   request.Amount,
			Status:   payment.StatusPending,
			Provider: providerName,
		},
		RedirectURL: strings.TrimRight(g.payURL, "/") + "/" + url.PathEscape(response.TransID),
	}, nil
}

func validatePurchase(request payment.PurchaseRequest) (int64, error) {
	if strings.TrimSpace(request.OrderID) == "" {
		return 0, fmt.Errorf("%w: order ID is required", payment.ErrInvalidRequest)
	}
	callback, err := url.ParseRequestURI(request.CallbackURL)
	if err != nil || callback.Host == "" || (callback.Scheme != "http" && callback.Scheme != "https") {
		return 0, fmt.Errorf("%w: callback URL must be an absolute HTTP(S) URL", payment.ErrInvalidRequest)
	}
	return normalizeAmount(request.Amount)
}

func normalizeAmount(money payment.Money) (int64, error) {
	amount := money.Amount
	switch strings.ToUpper(string(money.Currency)) {
	case string(payment.CurrencyIRR):
		if amount%10 != 0 {
			return 0, fmt.Errorf("%w: IRR amount must convert to whole toman", payment.ErrInvalidRequest)
		}
		amount /= 10
	case string(payment.CurrencyIRT):
	default:
		return 0, fmt.Errorf("%w: Aghaye Pardakht requires IRR or IRT", payment.ErrInvalidRequest)
	}
	if amount < minimumAmount || amount > maximumAmount {
		return 0, fmt.Errorf("%w: amount must be between %d and %d toman", payment.ErrInvalidRequest, minimumAmount, maximumAmount)
	}
	return amount, nil
}
