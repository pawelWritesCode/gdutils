// Package schema holds services that allows to validate JSON string against a schema.
//
// Package contains two types of JSON schema validators:
//
// raw - which accepts JSON schema string,
// reference - which accepts reference to JSON schema,
//
//	JSONSchemaRawXGValidator has ability to validate JSON schema string written with draft v4 v6 or v7.
//	JSONSchemaReferenceXGValidator has ability to validate JSON schema passed as URL or OS path written with draft v4 v6 or v7.
//
// By default, gojsonschema will try to detect the draft of a schema by using the $schema keyword and parse it
// in a strict draft-04, draft-06 or draft-07 mode. If $schema is missing, or the draft version is not explicitely set,
// a hybrid mode is used which merges together functionality of all drafts into one mode.
//
//	JSONSchemaRawQIValidator has ability to validate JSON schema string written with draft 7 & 2019-09
package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/qri-io/jsonschema"
	jschema "github.com/xeipuuv/gojsonschema"

	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/osutils"
	v "github.com/pawelWritesCode/gdutils/pkg/validator"
)

// JSONSchemaReferenceXGValidator is entity that has ability to validate data against JSON schema passed as reference.
// xeipuuv/gojsonschema is used under the hood.
type JSONSchemaReferenceXGValidator struct {
	fileValidator v.Validator
	urlValidator  v.Validator

	// schemasDir represents absolute path to JSON schemas directory.
	schemasDir string
}

// JSONSchemaRawXGValidator is entity that has ability to validate data against JSON schema passed as string.
// xeipuuv/gojsonschema is used under the hood
type JSONSchemaRawXGValidator struct{}

// JSONSchemaRawQIValidator is entity that has ability to validate data against JSON schema passed as string
// qri-io/jsonschema is used under the hood
type JSONSchemaRawQIValidator struct{}

// NewDefaultJSONSchemaReferenceXGValidator creates new JSONSchemaReferenceXGValidator with fixed services
func NewDefaultJSONSchemaReferenceXGValidator(schemasDir string) JSONSchemaReferenceXGValidator {
	return NewJSONSchemaReferenceXGValidator(schemasDir, osutils.NewFileValidator(), httpctx.NewURLValidator())
}

// NewJSONSchemaReferenceXGValidator creates new JSONSchemaReferenceXGValidator with provided services
func NewJSONSchemaReferenceXGValidator(schemasDir string, fileValidator v.Validator, urlValidator v.Validator) JSONSchemaReferenceXGValidator {
	return JSONSchemaReferenceXGValidator{
		fileValidator: fileValidator,
		urlValidator:  urlValidator,
		schemasDir:    schemasDir,
	}
}

// NewJSONSchemaRawXGValidator creates new JSONSchemaRawXGValidator
func NewJSONSchemaRawXGValidator() JSONSchemaRawXGValidator {
	return JSONSchemaRawXGValidator{}
}

// Validate validates document against JSON schema located in schemaPath.
// schemaPath may be URL or relative/full path to json schema on user OS
// according to xeipuuv/gojsonschema library it covers JSON Schema, draft v4 v6 & v7
func (jsv JSONSchemaReferenceXGValidator) Validate(document, schemaPath string) error {
	source, err := getSource(jsv.urlValidator, jsv.fileValidator, jsv.schemasDir, schemaPath)
	if err != nil {
		return err
	}

	result, err := jschema.Validate(jschema.NewReferenceLoader(source), jschema.NewStringLoader(document))
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

// Validate validates document against jsonSchema.
// according to xeipuuv/gojsonschema library it covers JSON Schema, draft v4 v6 & v7
func (j JSONSchemaRawXGValidator) Validate(document, jsonSchema string) error {
	result, err := jschema.Validate(jschema.NewStringLoader(jsonSchema), jschema.NewStringLoader(document))
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

// Validate validates document against json schema.
// according to library documentation it covers https://json-schema.org drafts 7 & 2019-09
func (j JSONSchemaRawQIValidator) Validate(document, jsonSchema string) error {
	rs := &jsonschema.Schema{}
	if err := json.Unmarshal([]byte(jsonSchema), rs); err != nil {
		return err
	}

	errs, err := rs.ValidateBytes(context.Background(), []byte(document))
	if err != nil {
		return err
	}

	var errStr string
	if len(errs) > 0 {
		for _, e := range errs {
			errStr += e.Error() + " "
		}

		err = errors.New(errStr)
	}

	return err
}

// getSource accepts rawSource, validate it and returns valid source
// available sources are: file system os path and URL
func getSource(urlValidator, fileValidator v.Validator, schemasDir, rawSource string) (string, error) {
	if rawSource == "" {
		return rawSource, errors.New("provided rawSource should not be empty string")
	}

	errURL := urlValidator.Validate(rawSource)
	if errURL == nil { // is valid URL
		return rawSource, nil
	}

	var pth string

	if path.IsAbs(rawSource) { // rawSource is valid absolute path
		pth = rawSource
	} else {
		pth = path.Clean(path.Join(schemasDir, rawSource))
	}

	errPath := fileValidator.Validate(pth)
	if errPath == nil { // pth points at some resource in user OS
		return fmt.Sprintf("%s%s", "file://", pth), nil
	}

	return "", fmt.Errorf("%s isn't valid path to any resource on your OS, nor valid URL", rawSource)
}
