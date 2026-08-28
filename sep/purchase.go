package sep

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/codenaline/payment"
)

const providerName = "sep"

type tokenRequest struct {
	Action      string `json:"Action"`
	TerminalID  int64  `json:"TerminalId"`
	Amount      int64  `json:"Amount"`
	ResNum      string `json:"ResNum"`
	RedirectURL string `json:"RedirectUrl"`
}

type tokenResponse struct {
	Status    int    `json:"status"`
	Token     string `json:"token"`
	ErrorCode string `json:"errorCode"`
	ErrorDesc string `json:"errorDesc"`
}

// Purchase obtains an SEP payment token.
func (g *Gateway) Purchase(ctx context.Context, request payment.PurchaseRequest) (payment.PurchaseResponse, error) {
	if err := validatePurchase(request); err != nil {
		return payment.PurchaseResponse{}, err
	}

	payload := tokenRequest{
		Action:      "Token",
		TerminalID:  g.terminalID,
		Amount:      request.Amount.Amount,
		ResNum:      request.OrderID,
		RedirectURL: request.CallbackURL,
	}
	var response tokenResponse
	if err := g.post(ctx, g.tokenURL, payload, &response); err != nil {
		return payment.PurchaseResponse{}, newError("purchase", 0, "request failed", err)
	}
	if response.Status != 1 || strings.TrimSpace(response.Token) == "" {
		code, _ := strconv.Atoi(response.ErrorCode)
		return payment.PurchaseResponse{}, newError("purchase", code, response.ErrorDesc, nil)
	}

	redirectURL, err := url.Parse(g.payURL)
	if err != nil {
		return payment.PurchaseResponse{}, newError("purchase", 0, "build redirect URL", fmt.Errorf("%w: %w", payment.ErrProvider, err))
	}
	query := redirectURL.Query()
	query.Set("token", response.Token)
	redirectURL.RawQuery = query.Encode()

	return payment.PurchaseResponse{
		Transaction: payment.Transaction{
			ID:       response.Token,
			Amount:   request.Amount,
			Status:   payment.StatusPending,
			Provider: providerName,
		},
		RedirectURL: redirectURL.String(),
	}, nil
}

func validatePurchase(request payment.PurchaseRequest) error {
	if strings.TrimSpace(request.OrderID) == "" {
		return fmt.Errorf("%w: order ID is required", payment.ErrInvalidRequest)
	}
	if request.Amount.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", payment.ErrInvalidRequest)
	}
	if !strings.EqualFold(string(request.Amount.Currency), string(payment.CurrencyIRR)) {
		return fmt.Errorf("%w: SEP requires IRR", payment.ErrInvalidRequest)
	}
	callback, err := url.ParseRequestURI(request.CallbackURL)
	if err != nil || callback.Host == "" || (callback.Scheme != "http" && callback.Scheme != "https") {
		return fmt.Errorf("%w: callback URL must be an absolute HTTP(S) URL", payment.ErrInvalidRequest)
	}
	return nil
}
