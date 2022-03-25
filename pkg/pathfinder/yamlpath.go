package pathfinder

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

type GoccyGoYamlFinder struct{}

func NewGoccyGoYamlFinder() GoccyGoYamlFinder {
	return GoccyGoYamlFinder{}
}

func (g GoccyGoYamlFinder) Find(expr string, jsonBytes []byte) (any, error) {
	yamlPath, err := yaml.PathString(expr)
	if err != nil {
		return nil, err
	}

	var result any
	err = yamlPath.Read(bytes.NewReader(jsonBytes), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
