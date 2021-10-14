package template

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

//ErrMissingKey occurs when storage is missing key
var ErrMissingKey = errors.New("missing entry for a key")

//TemplateEngine is entity that has ability to work with templated values
type TemplateEngine interface {
	//Replace replaces template values on that provided in storage
	Replace(templateValue string, storage map[string]interface{}) (string, error)
}

//TemplateManager is entity that has ability to manage templates
type TemplateManager struct{}

func New() TemplateManager {
	return TemplateManager{}
}

//Replace replaces template values in string on that ones provided in storage
//template values should exist between two brackets {{ }} preceded with dot, for example: "my name is: {{.NAME}}"
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
		return "", fmt.Errorf("%w, at least one of provided template values is not present in values storage", ErrMissingKey)
	}

	return strVal, nil
}
