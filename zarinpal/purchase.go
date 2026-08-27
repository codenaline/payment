package zarinpal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/codenaline/payment"
)

const providerName = "zarinpal"

// Purchase initiates a payment through Zarinpal.
func (g *Gateway) Purchase(ctx context.Context, request payment.PurchaseRequest) (payment.PurchaseResponse, error) {
	if err := validatePurchase(request); err != nil {
		return payment.PurchaseResponse{}, err
	}

	payload := requestPayload{
		MerchantID:  g.merchantID,
		Amount:      request.Amount.Amount,
		CallbackURL: request.CallbackURL,
		Description: request.Description,
		Metadata:    request.Metadata,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return payment.PurchaseResponse{}, purchaseError(0, "encode request", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, g.apiBaseURL+"/request.json", bytes.NewReader(body))
	if err != nil {
		return payment.PurchaseResponse{}, purchaseError(0, "create request", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	var response requestResponse
	if err := g.post(httpRequest, &response); err != nil {
		return payment.PurchaseResponse{}, purchaseError(0, "request failed", err)
	}
	if response.Data.Code != 100 || response.Data.Authority == "" {
		apiErr := decodeAPIError(response.Errors)
		if apiErr.Code == 0 {
			apiErr.Code = response.Data.Code
			apiErr.Message = response.Data.Message
		}
		return payment.PurchaseResponse{}, purchaseError(apiErr.Code, apiErr.Message, nil)
	}

	transaction := payment.Transaction{
		ID:       response.Data.Authority,
		Amount:   request.Amount,
		Status:   payment.StatusPending,
		Provider: providerName,
	}
	return payment.PurchaseResponse{
		Transaction: transaction,
		RedirectURL: g.payBaseURL + "/" + url.PathEscape(response.Data.Authority),
	}, nil
}

func validatePurchase(request payment.PurchaseRequest) error {
	if request.Amount.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", payment.ErrInvalidRequest)
	}
	if !strings.EqualFold(request.Amount.Currency, "IRR") {
		return fmt.Errorf("%w: Zarinpal requires IRR", payment.ErrInvalidRequest)
	}
	callback, err := url.ParseRequestURI(request.CallbackURL)
	if err != nil || callback.Scheme == "" || callback.Host == "" {
		return fmt.Errorf("%w: callback URL must be absolute", payment.ErrInvalidRequest)
	}
	if strings.TrimSpace(request.Description) == "" {
		return fmt.Errorf("%w: description is required", payment.ErrInvalidRequest)
	}
	return nil
}

func purchaseError(code int, message string, cause error) error {
	return &Error{Operation: "purchase", Code: code, Message: message, Kind: payment.ErrPurchaseFailed, Cause: cause}
}
