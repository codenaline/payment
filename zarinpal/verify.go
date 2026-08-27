package zarinpal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

// Verify verifies a payment through Zarinpal.
func (g *Gateway) Verify(ctx context.Context, request payment.VerifyRequest) (payment.Transaction, error) {
	if err := validateVerify(request); err != nil {
		return payment.Transaction{}, err
	}

	payload := verifyPayload{
		MerchantID: g.merchantID,
		Amount:     request.Amount.Amount,
		Authority:  request.TransactionID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return payment.Transaction{}, verificationError(0, "encode request", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, g.apiBaseURL+"/verify.json", bytes.NewReader(body))
	if err != nil {
		return payment.Transaction{}, verificationError(0, "create request", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	var response verifyResponse
	if err := g.post(httpRequest, &response); err != nil {
		return payment.Transaction{}, verificationError(0, "request failed", err)
	}
	if response.Data.Code != 100 && response.Data.Code != 101 {
		apiErr := decodeAPIError(response.Errors)
		if apiErr.Code == 0 {
			apiErr.Code = response.Data.Code
			apiErr.Message = response.Data.Message
		}
		return payment.Transaction{}, verificationError(apiErr.Code, apiErr.Message, nil)
	}

	return payment.Transaction{
		ID:          request.TransactionID,
		Amount:      request.Amount,
		Status:      payment.StatusPaid,
		ReferenceID: strconv.FormatInt(response.Data.RefID, 10),
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
	if !strings.EqualFold(request.Amount.Currency, "IRR") {
		return fmt.Errorf("%w: Zarinpal requires IRR", payment.ErrInvalidRequest)
	}
	return nil
}

func verificationError(code int, message string, cause error) error {
	return &payment.Error{Provider: providerName, Operation: "verification", Code: providerErrorCode(code), Message: message, Kind: payment.ErrVerificationFailed, Cause: cause}
}

var _ payment.Gateway = (*Gateway)(nil)
