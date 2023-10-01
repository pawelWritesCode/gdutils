// Package template holds utilities for working with templates.
package template

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

var (
	// ErrMissingStorage represents error when storage with data is missing
	ErrMissingStorage = errors.New("missing storage")

	// ErrMissingStorageValue represents error when storage doesn't have required data in it
	ErrMissingStorageValue = errors.New("missing storage value")
)

// TemplateManager is entity that has ability to manage templates.
type TemplateManager struct{}

func New() TemplateManager {
	return TemplateManager{}
}

// Replace replaces template values using provided storage.
// templateValue should exist between two brackets {{ }} preceded with dot, for example: "my name is: {{.NAME}}".
func (tm TemplateManager) Replace(templateValue string, storage map[string]any) (string, error) {
	if storage == nil {
		return "", fmt.Errorf("%w: passed nil storage for TemplateManager, storage should not be nil", ErrMissingStorage)
	}

	templ := template.Must(template.New("abc").Parse(templateValue))
	var buff bytes.Buffer
	err := templ.Execute(&buff, storage)
	if err != nil {
		return "", err
	}

	strVal := buff.String()

	if strings.Contains(strVal, "<no value>") {
		return "", fmt.Errorf("%w: string contains references to template values that are not present in provided storage", ErrMissingStorageValue)
	}

	return strVal, nil
}
