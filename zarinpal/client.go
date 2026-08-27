package zarinpal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const maxResponseSize = 1 << 20

func (g *Gateway) post(request *http.Request, result any) error {
	response, err := g.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize))
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	return nil
}
