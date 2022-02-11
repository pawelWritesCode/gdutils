// Package formatter holds utilities for working with different data formats.
package formatter

import (
	"encoding/json"
	"errors"

	"gopkg.in/yaml.v2"
)

// Deserializer describes ability to deserialize []byte in given format into v
type Deserializer interface {
	// Deserialize deserializes data on v
	Deserialize(data []byte, v interface{}) error
}

// JSONFormatter is entity that has ability to deserialize data in JSON format
type JSONFormatter struct{}

// YAMLFormatter is entity that has ability to deserialize data in YAML format
type YAMLFormatter struct{}

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

func NewAwareFormatter(JSONFormatter JSONFormatter, YAMLFormatter YAMLFormatter) AwareFormatter {
	return AwareFormatter{JSONFormatter: JSONFormatter, YAMLFormatter: YAMLFormatter}
}

// Deserialize data in format of JSON on v
func (J JSONFormatter) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
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
