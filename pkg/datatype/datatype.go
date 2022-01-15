// Package datatype holds utilities for working with different data formats.
package datatype

import (
	"encoding/json"
	"fmt"
)

// IsJSON checks whether bytes are in JSON format.
func IsJSON(bytes []byte) error {
	var js json.RawMessage
	err := json.Unmarshal(bytes, &js)
	if err != nil {
		return fmt.Errorf("expected last HTTP(s) response body to be in JSON format, err: %s", err.Error())
	}

	return nil
}
