// Package formatter holds utilities for working with different data formats.
package formatter

import (
	"encoding/json"
	"encoding/xml"
	"errors"

	"gopkg.in/yaml.v2"
)

// Formatter describes ability to serialize and deserialize data
type Formatter interface {
	// Deserialize deserializes data on v
	Deserialize(data []byte, v interface{}) error

	// Serialize serializes v
	Serialize(v interface{}) ([]byte, error)
}

// JSONFormatter is entity that has ability to work with JSON format
type JSONFormatter struct{}

// YAMLFormatter is entity that has ability to work with YAML format
type YAMLFormatter struct{}

// XMLFormatter is entity that has ability to work with XML format
type XMLFormatter struct{}

// AwareFormatter is entity that has ability to deserialize data in JSON or YAML format
type AwareFormatter struct {
	JSONFormatter JSONFormatter
	YAMLFormatter YAMLFormatter
}

func NewJSONFormatter() JSONFormatter {
	return JSONFormatter{}
}

func NewYAMLFormatter() YAMLFormatter {
	return YAMLFormatter{}
}

func NewXMLFormatter() XMLFormatter {
	return XMLFormatter{}
}

func NewAwareFormatter(JSONFormatter JSONFormatter, YAMLFormatter YAMLFormatter) AwareFormatter {
	return AwareFormatter{JSONFormatter: JSONFormatter, YAMLFormatter: YAMLFormatter}
}

// Deserialize data in format of JSON on v
func (J JSONFormatter) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Serialize serializes v into JSON format.
func (J JSONFormatter) Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize data in format of YAML on v
func (Y YAMLFormatter) Deserialize(data []byte, v interface{}) error {
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
func (y YAMLFormatter) Serialize(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Deserialize data in format of JSON or YAML on v
func (a AwareFormatter) Deserialize(data []byte, v interface{}) error {
	if err := a.JSONFormatter.Deserialize(data, v); err == nil {
		return nil
	}

	if err := a.YAMLFormatter.Deserialize(data, v); err == nil {
		return nil
	}

	return errors.New("could not deserialize on any of: json, xml, yaml")
}

// Deserialize data in format of XML on v.
func (X XMLFormatter) Deserialize(data []byte, v interface{}) error {
	if data == nil {
		return errors.New("data should not be nil")
	}

	if len(data) == 0 {
		return errors.New("data should not be empty []byte()")
	}

	return xml.Unmarshal(data, v)
}

// Serialize serializes v into XML format.
func (X XMLFormatter) Serialize(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}
