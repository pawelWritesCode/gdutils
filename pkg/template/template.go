// Package template holds utilities for working with templates.
package template

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
)

// Engine is entity that has ability to work with templates.
type Engine interface {
	// Replace replaces template values using provided storage.
	Replace(templateValue string, storage map[string]interface{}) (string, error)
}

// TemplateManager is entity that has ability to manage templates.
type TemplateManager struct{}

func New() TemplateManager {
	return TemplateManager{}
}

// Replace replaces template values using provided storage.
// templateValue should exist between two brackets {{ }} preceded with dot, for example: "my name is: {{.NAME}}".
func (tm TemplateManager) Replace(templateValue string, storage map[string]interface{}) (string, error) {
	if storage == nil {
		return "", errors.New("missing values storage for template manager")
	}

	templ := template.Must(template.New("abc").Parse(templateValue))
	var buff bytes.Buffer
	err := templ.Execute(&buff, storage)
	if err != nil {
		return "", err
	}

	strVal := buff.String()

	if strings.Contains(strVal, "<no value>") {
		return "", fmt.Errorf("%w, at least one of provided values is not present in storage", cache.ErrMissingKey)
	}

	return strVal, nil
}
