// Package pathfinder holds utilities for working with JSON path.
package pathfinder

import (
	"encoding/json"
	"fmt"

	"github.com/oliveagle/jsonpath"
	"github.com/pawelWritesCode/qjson"
	"github.com/tidwall/gjson"
)

// QJSONFinder represents implementation of JSON path from https://github.com/pawelWritesCode/qjson library
type QJSONFinder struct{}

// OliveagleJSONFinder represents implementation of JSON path from https://github.com/oliveagle/jsonpath library
type OliveagleJSONFinder struct{}

// GJSONFinder represents implementation of JSON path from https://github.com/tidwall/gjson library
type GJSONFinder struct{}

// DynamicJSONPathFinder is entity that has ability to obtain data from JSON from given expression.
// Entity knows how to determine whether expression matches
// https://github.com/tidwall/gjson or https://github.com/oliveagle/jsonpath syntax
type DynamicJSONPathFinder struct {
	gjson               GJSONFinder
	oliveagleJSONFinder OliveagleJSONFinder
}

func NewQJSONFinder() QJSONFinder {
	return QJSONFinder{}
}

func NewOliveagleJSONFinder() OliveagleJSONFinder {
	return OliveagleJSONFinder{}
}

func NewGJSONFinder() GJSONFinder {
	return GJSONFinder{}
}

func NewDynamicJSONPathFinder(gjson GJSONFinder, oliveagleJSONFinder OliveagleJSONFinder) *DynamicJSONPathFinder {
	return &DynamicJSONPathFinder{gjson: gjson, oliveagleJSONFinder: oliveagleJSONFinder}
}

// Find obtains data from jsonBytes according to given expr valid with pawelWritesCode/qjson library
func (Q QJSONFinder) Find(expr string, jsonBytes []byte) (any, error) {
	return qjson.Resolve(expr, jsonBytes)
}

// Find obtains data from jsonBytes according to given expr valid with oliveagle/jsonpath library
func (o OliveagleJSONFinder) Find(expr string, jsonBytes []byte) (any, error) {
	var jsonData any
	err := json.Unmarshal(jsonBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonpath.JsonPathLookup(jsonData, expr)
}

// Find obtains data from jsonBytes according to given expr.
// It accepts expr in format acceptable by tidwall/gjson or oliveagle/jsonpath libraries.
func (d DynamicJSONPathFinder) Find(expr string, jsonBytes []byte) (any, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("json path can't be empty string")
	}

	if expr[0:1] == "$" {
		return d.oliveagleJSONFinder.Find(expr, jsonBytes)
	}

	return d.gjson.Find(expr, jsonBytes)
}

// Find obtains data from jsonBytes according to given expr valid with tidwall/gjson library
func (G GJSONFinder) Find(expr string, bytes []byte) (any, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("provided empty expression")
	}

	if !gjson.ValidBytes(bytes) {
		return nil, fmt.Errorf("detected invalid JSON")
	}

	result := gjson.GetBytes(bytes, expr)

	if !result.Exists() {
		return nil, fmt.Errorf("could not find node, using expression %s", expr)
	}

	return result.Value(), nil
}
