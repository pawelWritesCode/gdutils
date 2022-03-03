// Package pathfinder holds utilities for working with JSON path.
package pathfinder

import (
	"encoding/json"
	"fmt"

	"github.com/oliveagle/jsonpath"
	"github.com/pawelWritesCode/qjson"
)

// QJSONFinder represents implementation of JSON path from https://github.com/pawelWritesCode/qjson library
type QJSONFinder struct{}

// OliveagleJSONFinder represents implementation of JSON path from https://github.com/oliveagle/jsonpath library
type OliveagleJSONFinder struct{}

// DynamicJSONPathFinder is entity that has ability to obtain data from JSON from given expression.
// Entity knows how to determine whether expression matches
// https://github.com/pawelWritesCode/qjson or https://github.com/oliveagle/jsonpath syntax
type DynamicJSONPathFinder struct {
	qjson               QJSONFinder
	oliveagleJSONFinder OliveagleJSONFinder
}

func NewQJSONFinder() QJSONFinder {
	return QJSONFinder{}
}

func NewOliveagleJSONFinder() OliveagleJSONFinder {
	return OliveagleJSONFinder{}
}

func NewDynamicJSONPathFinder(qjson QJSONFinder, oliveagleJSONFinder OliveagleJSONFinder) *DynamicJSONPathFinder {
	return &DynamicJSONPathFinder{qjson: qjson, oliveagleJSONFinder: oliveagleJSONFinder}
}

// Find obtains data from jsonBytes according to given expr valid with oliveagle/jsonpath library
func (Q QJSONFinder) Find(expr string, jsonBytes []byte) (interface{}, error) {
	return qjson.Resolve(expr, jsonBytes)
}

// Find obtains data from jsonBytes according to given expr valid with pawelWritesCode/qjson library
func (o OliveagleJSONFinder) Find(expr string, jsonBytes []byte) (interface{}, error) {
	var jsonData interface{}
	err := json.Unmarshal(jsonBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonpath.JsonPathLookup(interface{}(jsonData), expr)
}

// Find obtains data from jsonBytes according to given expr.
// It accepts expr in format acceptable by pawelWritesCode/qjson or oliveagle/jsonpath libraries.
func (d DynamicJSONPathFinder) Find(expr string, jsonBytes []byte) (interface{}, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("json path can't be empty string")
	}

	if expr[0:1] == "$" {
		return d.oliveagleJSONFinder.Find(expr, jsonBytes)
	}

	return d.qjson.Find(expr, jsonBytes)
}
