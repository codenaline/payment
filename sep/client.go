package sep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/codenaline/payment"
)

const maxResponseSize = 1 << 20

func (g *Gateway) post(ctx context.Context, endpoint string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: encode request: %w", payment.ErrProvider, err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: create request: %w", payment.ErrProvider, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := g.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %w", payment.ErrNetwork, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: unexpected HTTP status %s", payment.ErrProvider, response.Status)
	}
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("%w: read response: %w", payment.ErrNetwork, err)
	}
	if len(responseBody) > maxResponseSize {
		return fmt.Errorf("%w: response exceeds %d bytes", payment.ErrProvider, maxResponseSize)
	}
	if err := json.Unmarshal(responseBody, result); err != nil {
		return fmt.Errorf("%w: decode response: %w", payment.ErrProvider, err)
	}
	return nil
}
