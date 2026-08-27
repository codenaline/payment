package zarinpal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/codenaline/payment"
)

const maxResponseSize = 1 << 20

func (g *Gateway) post(request *http.Request, result any) error {
	response, err := g.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %w", payment.ErrNetwork, err)
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize))
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("%w: decode response: %v", payment.ErrProvider, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: unexpected HTTP status %s", payment.ErrProvider, response.Status)
	}
	return nil
}
