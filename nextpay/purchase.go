package nextpay

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

const providerName = "nextpay"

// Purchase creates a NextPay transaction token.
func (g *Gateway) Purchase(ctx context.Context, request payment.PurchaseRequest) (payment.PurchaseResponse, error) {
	if err := validatePurchase(request); err != nil {
		return payment.PurchaseResponse{}, err
	}

	form := url.Values{
		"api_key":      {g.apiKey},
		"order_id":     {request.OrderID},
		"amount":       {strconv.FormatInt(request.Amount.Amount, 10)},
		"callback_uri": {request.CallbackURL},
		"currency":     {strings.ToUpper(request.Amount.Currency)},
	}
	if request.Description != "" {
		form.Set("payer_desc", request.Description)
	}
	for _, field := range []string{"customer_phone", "payer_name", "auto_verify", "allowed_card"} {
		if value := request.Metadata[field]; value != "" {
			form.Set(field, value)
		}
	}

	var response tokenResponse
	if err := g.post(ctx, "/token", form, &response); err != nil {
		return payment.PurchaseResponse{}, newError("purchase", 0, "request failed", err)
	}
	if response.Code != int(CodeTokenCreated) || response.TransID == "" {
		return payment.PurchaseResponse{}, newError("purchase", response.Code, message(Code(response.Code)), nil)
	}

	transaction := payment.Transaction{
		ID:       response.TransID,
		Amount:   request.Amount,
		Status:   payment.StatusPending,
		Provider: providerName,
	}
	return payment.PurchaseResponse{
		Transaction: transaction,
		RedirectURL: g.payBaseURL + "/" + url.PathEscape(response.TransID),
	}, nil
}

func validatePurchase(request payment.PurchaseRequest) error {
	if strings.TrimSpace(request.OrderID) == "" {
		return fmt.Errorf("%w: order ID is required", payment.ErrInvalidRequest)
	}
	if request.Amount.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", payment.ErrInvalidRequest)
	}
	currency := strings.ToUpper(request.Amount.Currency)
	if currency != "IRR" && currency != "IRT" {
		return fmt.Errorf("%w: NextPay requires IRR or IRT", payment.ErrInvalidRequest)
	}
	callback, err := url.ParseRequestURI(request.CallbackURL)
	if err != nil || callback.Scheme == "" || callback.Host == "" {
		return fmt.Errorf("%w: callback URL must be absolute", payment.ErrInvalidRequest)
	}
	return nil
}

func message(code Code) string {
	switch code {
	case CodePaymentRejected:
		return "payment rejected by payer or bank"
	case CodePaymentPending:
		return "payment is waiting for the bank"
	case CodePaymentCanceled:
		return "payment canceled"
	case CodeInvalidAmount:
		return "invalid amount"
	case CodeInvalidAPIKey:
		return "invalid API key"
	case CodeInvalidTransaction:
		return "invalid transaction token"
	case CodeTransactionNotFound:
		return "transaction not found"
	case CodeSystemError:
		return "NextPay system error"
	case CodeRefundFailed:
		return "refund failed"
	default:
		return "NextPay request failed"
	}
}
