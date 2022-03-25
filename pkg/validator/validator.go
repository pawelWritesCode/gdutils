// Package validator holds utilities for validating data.
package validator

// SchemaValidator describes entity that can validate document against some kind of schema.
type SchemaValidator interface {
	// Validate validates document against some kind of schema located in schemaPath.
	Validate(document, schemaPath string) error
}

// Validator describes validator
type Validator interface {
	// Validate validates in
	Validate(in any) error
}
