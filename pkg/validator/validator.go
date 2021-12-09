package validator

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator describes entity that can validate something
type SchemaValidator interface {
	// Validate validates something
	Validate(document, schemaPath string) error
}

//JSONSchemaValidator is entity that has ability to validate data against JSON schema
type JSONSchemaValidator struct {
	//schemasDir represents absolute path to JSON schemas directory
	schemasDir string
}

func NewJSONSchemaValidator(schemasDir string) JSONSchemaValidator {
	return JSONSchemaValidator{schemasDir: schemasDir}
}

func (jsv JSONSchemaValidator) Validate(document, schemaPath string) error {
	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s/%s", jsv.schemasDir, schemaPath))
	documentLoader := gojsonschema.NewStringLoader(document)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errSum := ""
		for _, err := range result.Errors() {
			errSum += err.String()
		}

		return errors.New(errSum)
	}

	return nil
}
