// Package serializer holds utilities for working with different data formats.
package serializer

import (
	"encoding/json"
	"encoding/xml"
	"errors"

	"gopkg.in/yaml.v2"
)

// JSON is entity that has ability to work with JSON format
type JSON struct{}

// YAML is entity that has ability to work with YAML format
type YAML struct{}

// XML is entity that has ability to work with XML format
type XML struct{}

func NewJSONFormatter() JSON {
	return JSON{}
}

func NewYAMLFormatter() YAML {
	return YAML{}
}

func NewXMLFormatter() XML {
	return XML{}
}

// Deserialize deserializes data in JSON format on v.
func (serializer JSON) Deserialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// Serialize serializes v into JSON format.
func (serializer JSON) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize deserializes data in YAML format on v.
func (serializer YAML) Deserialize(data []byte, v any) error {
	if data == nil {
		return errors.New("data should not be nil")
	}

	if len(data) == 0 {
		return errors.New("data should not be empty []byte()")
	}

	if err := json.Unmarshal(data, v); err == nil {
		return errors.New("data is in JSON format, expected YAML")
	}

	return yaml.UnmarshalStrict(data, v)
}

// Serialize serializes v into YAML format.
func (serializer YAML) Serialize(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

// Deserialize deserializes data in XML format on v.
func (serializer XML) Deserialize(data []byte, v any) error {
	if data == nil {
		return errors.New("data should not be nil")
	}

	if len(data) == 0 {
		return errors.New("data should not be empty []byte()")
	}

	return xml.Unmarshal(data, v)
}

// Serialize serializes v into XML format.
func (serializer XML) Serialize(v any) ([]byte, error) {
	return xml.Marshal(v)
}
