package pathfinder

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

type GoccyGoYamlFinder struct{}

func NewGoccyGoYamlFinder() GoccyGoYamlFinder {
	return GoccyGoYamlFinder{}
}

func (g GoccyGoYamlFinder) Find(expr string, jsonBytes []byte) (interface{}, error) {
	yamlPath, err := yaml.PathString(expr)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = yamlPath.Read(bytes.NewReader(jsonBytes), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
