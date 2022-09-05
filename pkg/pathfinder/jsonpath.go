// Package pathfinder holds utilities for working with JSON path.
package pathfinder

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/antchfx/jsonquery"
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

// AntchfxJSONQueryFinder represents implementation of JSON path from https://github.com/antchfx/jsonquery library
type AntchfxJSONQueryFinder struct{}

// DynamicJSONPathFinder is entity that has ability to obtain data from JSON from given expression.
// Entity knows how to determine whether expression matches
// https://github.com/tidwall/gjson, https://github.com/oliveagle/jsonpath or https://github.com/antchfx/jsonquery syntax
type DynamicJSONPathFinder struct {
	gjson               GJSONFinder
	oliveagleJSONFinder OliveagleJSONFinder
	antchfxJSONQuery    AntchfxJSONQueryFinder
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

func NewAntchfxJSONQueryFinder() AntchfxJSONQueryFinder {
	return AntchfxJSONQueryFinder{}
}

func NewDynamicJSONPathFinder(gjson GJSONFinder, oliveagleJSONFinder OliveagleJSONFinder, antchfxJSONQuery AntchfxJSONQueryFinder) *DynamicJSONPathFinder {
	return &DynamicJSONPathFinder{gjson: gjson, oliveagleJSONFinder: oliveagleJSONFinder, antchfxJSONQuery: antchfxJSONQuery}
}

// Find obtains data from jsonBytes according to given expr valid with pawelWritesCode/qjson library
func (Q QJSONFinder) Find(expr string, jsonBytes []byte) (any, error) {
	return qjson.Resolve(expr, jsonBytes)
}

// Find obtains data from jsonBytes according to given expr.
// It accepts expr in format acceptable by tidwall/gjson, oliveagle/jsonpath or antchfx/jsonquery libraries.
func (d DynamicJSONPathFinder) Find(expr string, jsonBytes []byte) (any, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("json path can't be empty string")
	}

	if expr[0:1] == "$" {
		return d.oliveagleJSONFinder.Find(expr, jsonBytes)
	}

	if expr[0:1] == "/" {
		return d.antchfxJSONQuery.Find(expr, jsonBytes)
	}

	return d.gjson.Find(expr, jsonBytes)
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

// Find obtains data from jsonBytes according to given expr valid with tidwall/gjson library
func (G GJSONFinder) Find(expr string, b []byte) (any, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("provided empty expression")
	}

	if !gjson.ValidBytes(b) {
		return nil, fmt.Errorf("detected invalid JSON")
	}

	result := gjson.GetBytes(b, expr)

	if !result.Exists() {
		return nil, fmt.Errorf("could not find node, using expression %s", expr)
	}

	return result.Value(), nil
}

// Find obtains data from jsonBytes according to given expr valid with antchfx/jsonquery library
func (a AntchfxJSONQueryFinder) Find(expr string, b []byte) (any, error) {
	if len(expr) == 0 {
		return nil, fmt.Errorf("provided empty expression")
	}

	doc, err := jsonquery.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("detected invalid JSON")
	}

	nodes, err := jsonquery.QueryAll(doc, expr)
	if err != nil {
		return nil, fmt.Errorf("could not find node, using expression %s, err: %w", expr, err)
	}

	if nodes == nil {
		return nil, fmt.Errorf("could not find node, using expression %s", expr)
	}

	if len(nodes) == 1 {
		return any(nodes[0].Value()), nil
	}

	if len(nodes) > 1 {
		results := make([]any, 0, len(nodes))
		for _, node := range nodes {
			results = append(results, node.Value())
		}

		return results, nil
	}

	return nil, fmt.Errorf("could not find %s in given JSON bytes", expr)
}
