package sep

import (
	"context"
	"fmt"
	"strings"

	"github.com/codenaline/payment"
)

type verifyRequest struct {
	RefNum         string `json:"RefNum"`
	TerminalNumber int64  `json:"TerminalNumber"`
}

type verifyResponse struct {
	TransactionDetail transactionDetail `json:"TransactionDetail"`
	ResultCode        int               `json:"ResultCode"`
	ResultDescription string            `json:"ResultDescription"`
	Success           bool              `json:"Success"`
}

type transactionDetail struct {
	RRN             string `json:"RRN"`
	RefNum          string `json:"RefNum"`
	MaskedPan       string `json:"MaskedPan"`
	HashedPan       string `json:"HashedPan"`
	TerminalNumber  int64  `json:"TerminalNumber"`
	OrginalAmount   int64  `json:"OrginalAmount"`
	AffectiveAmount int64  `json:"AffectiveAmount"`
	StraceDate      string `json:"StraceDate"`
	StraceNo        string `json:"StraceNo"`
}

// Verify verifies an SEP digital receipt. TransactionID must be the RefNum
// returned by SEP to the application's callback URL.
func (g *Gateway) Verify(ctx context.Context, request payment.VerifyRequest) (payment.Transaction, error) {
	if err := validateVerify(request); err != nil {
		return payment.Transaction{}, err
	}

	payload := verifyRequest{
		RefNum:         request.TransactionID,
		TerminalNumber: g.terminalID,
	}
	var response verifyResponse
	if err := g.post(ctx, g.verifyURL, payload, &response); err != nil {
		return payment.Transaction{}, fmt.Errorf("SEP verification: %w", err)
	}
	if !response.Success || response.ResultCode != 0 {
		return payment.Transaction{}, fmt.Errorf("%w: SEP verification failed with code %d: %s", payment.ErrProvider, response.ResultCode, response.ResultDescription)
	}
	if response.TransactionDetail.RefNum != request.TransactionID {
		return payment.Transaction{}, fmt.Errorf("%w: SEP returned a different reference number", payment.ErrProvider)
	}
	if response.TransactionDetail.TerminalNumber != g.terminalID {
		return payment.Transaction{}, fmt.Errorf("%w: SEP returned a different terminal number", payment.ErrProvider)
	}
	if response.TransactionDetail.OrginalAmount != request.Amount.Amount {
		return payment.Transaction{}, fmt.Errorf("%w: SEP returned a different amount", payment.ErrProvider)
	}

	return payment.Transaction{
		ID:          request.TransactionID,
		Amount:      request.Amount,
		Status:      payment.StatusPaid,
		ReferenceID: response.TransactionDetail.RRN,
		Provider:    providerName,
	}, nil
}

func validateVerify(request payment.VerifyRequest) error {
	if strings.TrimSpace(request.TransactionID) == "" {
		return fmt.Errorf("%w: SEP reference number is required", payment.ErrInvalidRequest)
	}
	if request.Amount.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", payment.ErrInvalidRequest)
	}
	if !strings.EqualFold(string(request.Amount.Currency), string(payment.CurrencyIRR)) {
		return fmt.Errorf("%w: SEP requires IRR", payment.ErrInvalidRequest)
	}
	return nil
}

var _ payment.Gateway = (*Gateway)(nil)
