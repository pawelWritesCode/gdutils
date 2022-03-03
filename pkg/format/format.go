package format

import (
	"encoding/json"

	"github.com/goccy/go-yaml"
)

const (
	// JSON describes JSON data format.
	JSON DataFormat = "JSON"

	// YAML describes Yaml data format.
	YAML DataFormat = "YAML"

	// PlainText describes plan text data format.
	PlainText DataFormat = "plain text"
)

// DataFormat describes format of data.
type DataFormat string

// IsJSON checks whether bytes are in JSON format.
func IsJSON(bytes []byte) bool {
	var js json.RawMessage
	err := json.Unmarshal(bytes, &js)

	return err == nil
}

//IsYAML checks whether bytes are in YAML format.
func IsYAML(bytes []byte) bool {
	if IsJSON(bytes) {
		return false
	}

	var y map[string]interface{}
	err := yaml.Unmarshal(bytes, &y)
	return err == nil
}
