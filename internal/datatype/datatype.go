package datatype

import (
	"encoding/json"
	"fmt"
)

//IsJSON checks if bytes are in JSON format
func IsJSON(bytes []byte) error {
	var js json.RawMessage
	err := json.Unmarshal(bytes, &js)
	if err != nil {
		return fmt.Errorf("expected last HTTP(s) response body to be JSON, err: %s", err.Error())
	}

	return nil
}
