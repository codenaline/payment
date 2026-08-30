package aqayepardakht

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type apiResponse struct {
	Status  string `json:"status"`
	Code    Code   `json:"code"`
	TransID string `json:"transid,omitempty"`
}

func (c *Code) UnmarshalJSON(data []byte) error {
	value := string(bytes.TrimSpace(data))
	if value == "" || value == "null" {
		*c = 0
		return nil
	}
	if value[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		value = text
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid provider code %q", value)
	}
	*c = Code(number)
	return nil
}
