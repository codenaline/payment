package nextpay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/codenaline/payment"
)

const maxResponseSize = 1 << 20

func (g *Gateway) post(ctx context.Context, path string, form url.Values, result any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, g.apiBaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%w: create request: %w", payment.ErrProvider, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := g.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %w", payment.ErrNetwork, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: unexpected HTTP status %s", payment.ErrProvider, response.Status)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize))
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("%w: decode response: %w", payment.ErrProvider, err)
	}
	return nil
}
