package template

import (
	"bytes"
	"text/template"
)

//TemplateEngine is entity that has ability to replace template values
type TemplateEngine interface {
	//Replace replaces template values in string on provided in storage
	Replace(templateValue string, storage map[string]interface{}) (string, error)
}

//TemplateManager is entity that has ability to manage templates
type TemplateManager struct{}

func New() TemplateManager {
	return TemplateManager{}
}

//Replace replaces template values in string on provided in storage
//template values should exist between two brackets {{ }} preceded with dot, for example: "my name is: {{.NAME}}"
func (tm TemplateManager) Replace(templateValue string, storage map[string]interface{}) (string, error) {
	templ := template.Must(template.New("abc").Parse(templateValue))
	var buff bytes.Buffer
	err := templ.Execute(&buff, storage)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}
