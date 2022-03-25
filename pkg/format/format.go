package format

import (
	"bytes"
	"encoding/json"
	"encoding/xml"

	"github.com/goccy/go-yaml"
)

const (
	// JSON describes JSON data format.
	JSON DataFormat = "JSON"

	// YAML describes Yaml data format.
	YAML DataFormat = "YAML"

	// XML describes XML data format.
	XML DataFormat = "XML"

	// PlainText describes plan text data format.
	PlainText DataFormat = "plain text"
)

// DataFormat describes format of data.
type DataFormat string

// IsJSON checks whether bytes are in JSON format.
func IsJSON(b []byte) bool {
	var js json.RawMessage
	err := json.Unmarshal(b, &js)

	return err == nil
}

//IsYAML checks whether bytes are in YAML format.
func IsYAML(b []byte) bool {
	if IsJSON(b) {
		return false
	}

	if IsXML(b) {
		return false
	}

	// yaml.UnmarshalWithOptions parses successfully any plain text,
	// to detect text that is not in yaml format, we assume, there must be,
	// at least one key: value pair in yaml
	if !bytes.Contains(b, []byte(":")) {
		return false
	}

	var y any
	return yaml.UnmarshalWithOptions(b, &y, yaml.Strict()) == nil
}

// IsXML checks whether bytes are in XML format.
func IsXML(b []byte) bool {
	var v any
	err := xml.Unmarshal(b, &v)
	return err == nil
}
