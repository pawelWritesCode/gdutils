// Package jsonpath holds utilities for working with JSON path.
package jsonpath

import (
	"encoding/json"
	"fmt"

	"github.com/oliveagle/jsonpath"
	"github.com/pawelWritesCode/qjson"
)

// Resolver describes ability to obtain node(s) from JSON
type Resolver interface {
	// Resolve obtains data from jsonBytes according to given expression
	Resolve(expr string, jsonBytes []byte) (interface{}, error)
}

// DynamicJSONPathResolver is entity that has ability to obtain data from JSON from given expression.
// Entity knows how to determine whether expression matches
// https://github.com/pawelWritesCode/qjson or https://github.com/oliveagle/jsonpath syntax
type DynamicJSONPathResolver struct {
	qjson             QJSONResolver
	oliveagleJSONpath OliveagleJSONpath
}

// QJSONResolver represents implementation of JSON path from https://github.com/pawelWritesCode/qjson library
type QJSONResolver struct{}

// OliveagleJSONpath represents implementation of JSON path from https://github.com/oliveagle/jsonpath library
type OliveagleJSONpath struct{}

func NewDynamicJSONPathResolver(qjson QJSONResolver, oliveagleJSONpath OliveagleJSONpath) *DynamicJSONPathResolver {
	return &DynamicJSONPathResolver{qjson: qjson, oliveagleJSONpath: oliveagleJSONpath}
}

func NewQJSONResolver() QJSONResolver {
	return QJSONResolver{}
}

func NewOliveagleJSONpath() OliveagleJSONpath {
	return OliveagleJSONpath{}
}

// Resolve obtains data from jsonBytes according to given expr.
// It accepts expr in format acceptable by pawelWritesCode/qjson or oliveagle/jsonpath libraries.
func (d DynamicJSONPathResolver) Resolve(expr string, jsonBytes []byte) (interface{}, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("json path can't be empty string")
	}

	if expr[0:1] == "$" {
		return d.oliveagleJSONpath.Resolve(expr, jsonBytes)
	}

	return d.qjson.Resolve(expr, jsonBytes)
}

// Resolve obtains data from jsonBytes according to given expr valid with oliveagle/jsonpath library
func (Q QJSONResolver) Resolve(expr string, jsonBytes []byte) (interface{}, error) {
	return qjson.Resolve(expr, jsonBytes)
}

// Resolve obtains data from jsonBytes according to given expr valid with pawelWritesCode/qjson library
func (o OliveagleJSONpath) Resolve(expr string, jsonBytes []byte) (interface{}, error) {
	var jsonData interface{}
	err := json.Unmarshal(jsonBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonpath.JsonPathLookup(interface{}(jsonData), expr)
}
