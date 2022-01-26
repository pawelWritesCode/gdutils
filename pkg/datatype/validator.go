package datatype

import (
	"errors"
	"fmt"
	"path"

	jschema "github.com/xeipuuv/gojsonschema"

	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/osutils"
	v "github.com/pawelWritesCode/gdutils/pkg/validator"
)

// JSONSchemaValidator is entity that has ability to validate data against JSON schema.
type JSONSchemaValidator struct {
	fileValidator v.Validator
	urlValidator  v.Validator

	// schemasDir represents absolute path to JSON schemas directory.
	schemasDir string
}

func NewDefaultJSONSchemaValidator(schemasDir string) JSONSchemaValidator {
	return NewJSONSchemaValidator(schemasDir, osutils.NewFileValidator(), httpctx.NewURLValidator())
}

func NewJSONSchemaValidator(schemasDir string, fileValidator v.Validator, urlValidator v.Validator) JSONSchemaValidator {
	return JSONSchemaValidator{
		fileValidator: fileValidator,
		urlValidator:  urlValidator,
		schemasDir:    schemasDir,
	}
}

// Validate validates document against JSON schema located in schemaPath.
// schemaPath may be URL or relative/full path to json schema on user OS
func (jsv JSONSchemaValidator) Validate(document, schemaPath string) error {
	source, err := jsv.getSource(schemaPath)
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

// getSource accepts rawSource, validate it and returns valid source
// available sources are: file system os path and URL
func (jsv JSONSchemaValidator) getSource(rawSource string) (string, error) {
	if rawSource == "" {
		return rawSource, errors.New("provided rawSource should not be empty string")
	}

	errURL := jsv.urlValidator.Validate(rawSource)
	if errURL == nil { // is valid URL
		return rawSource, nil
	}

	var pth string

	if path.IsAbs(rawSource) { // rawSource is valid absolute path
		pth = rawSource
	} else {
		pth = path.Clean(path.Join(jsv.schemasDir, rawSource))
	}

	errPath := jsv.fileValidator.Validate(pth)
	if errPath == nil { // pth points at some resource in user OS
		return fmt.Sprintf("%s%s", "file://", pth), nil
	}

	return "", fmt.Errorf("%s isn't valid path to any resource on your OS, nor valid URL", rawSource)
}
