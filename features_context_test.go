package gdutils

import (
	"net/http"
	"testing"
)

func TestApiFeature_theJSONNodeShouldBeOfValue(t *testing.T) {
	type fields struct {
		saved            map[string]interface{}
		lastResponse     *http.Response
		lastResponseBody []byte
		baseUrl          string
	}
	type args struct {
		expr      string
		dataType  string
		dataValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "empty json", fields: fields{
			saved:            nil,
			lastResponse:     nil,
			lastResponseBody: nil,
			baseUrl:          "",
		}, args: args{
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: true},
		{name: "json with first level field with string data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"name": "ivo"
}`),
			baseUrl: "",
		}, args: args{
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: false},
		{name: "json with first level field with int data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"number": 10
}`),
			baseUrl: "",
		}, args: args{
			expr:      "number",
			dataType:  "int",
			dataValue: "10",
		}, wantErr: false},
		{name: "json with first level field with float64 data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"number": 10.1
}`),
			baseUrl: "",
		}, args: args{
			expr:      "number",
			dataType:  "float",
			dataValue: "10.1",
		}, wantErr: false},
		{name: "json with first level field with bool data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"is": true
}`),
			baseUrl: "",
		}, args: args{
			expr:      "is",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"data": {
		"name": "Is empty",
		"value": true
	}
}`),
			baseUrl: "",
		}, args: args{
			expr:      "data.value",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"data":	[
			{
				"name": "Is empty",
				"value": true
			},
			{
				"name": "Is big",
				"value": false
			}
		]
}`),
			baseUrl: "",
		}, args: args{
			expr:      "data[1].value",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`
{
	"data":	[
			true,
			false
		]
}`),
			baseUrl: "",
		}, args: args{
			expr:      "data[1]",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &ApiFeature{
				saved:            tt.fields.saved,
				lastResponse:     tt.fields.lastResponse,
				lastResponseBody: tt.fields.lastResponseBody,
			}
			if err := af.TheJSONNodeShouldBeOfValue(tt.args.expr, tt.args.dataType, tt.args.dataValue); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldBeSliceOfLength(t *testing.T) {
	type fields struct {
		saved            map[string]interface{}
		lastResponse     *http.Response
		lastResponseBody []byte
		baseUrl          string
	}
	type args struct {
		expr   string
		length int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "no resp body", fields: fields{
			saved:            nil,
			lastResponse:     nil,
			lastResponseBody: nil,
			baseUrl:          "",
		}, args: args{
			expr:   "anykey",
			length: 0,
		}, wantErr: true},
		{name: "key is not slice", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`{
	"name": "xyz"	
}`),
			baseUrl: "",
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is not slice #2", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`{
	"name": {
		"details": "xyz"
	}
}`),
			baseUrl: "",
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is slice but length does not match", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`{
	"names": ["a", "b"]
}`),
			baseUrl: "",
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is slice and length match", fields: fields{
			saved:        nil,
			lastResponse: nil,
			lastResponseBody: []byte(`{
	"names": ["a", "b"]
}`),
			baseUrl: "",
		}, args: args{
			expr:   "names",
			length: 2,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &ApiFeature{
				saved:            tt.fields.saved,
				lastResponse:     tt.fields.lastResponse,
				lastResponseBody: tt.fields.lastResponseBody,
			}
			if err := af.TheJSONNodeShouldBeSliceOfLength(tt.args.expr, tt.args.length); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeSliceOfLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheResponseShouldBeInXML(t *testing.T) {
	type fields struct {
		saved            map[string]interface{}
		lastResponse     *http.Response
		lastResponseBody []byte
		isDebug          bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "no data", fields: fields{
			lastResponseBody: nil,
		}, wantErr: true},
		{name: "no data", fields: fields{
			lastResponseBody: []byte(""),
		}, wantErr: true},
		{name: "json data", fields: fields{
			lastResponseBody: []byte(`{"user": "pawel"}`),
		}, wantErr: true},
		{name: "raw text data", fields: fields{
			lastResponseBody: []byte(`abc`),
		}, wantErr: true},
		{name: "xml data #1", fields: fields{
			lastResponseBody: []byte(`<data> xxx </data>`),
		}, wantErr: false},
		{name: "xml data #2", fields: fields{
			lastResponseBody: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Data>
	<Id>1</Id>
</Data>
`),
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &ApiFeature{
				saved:            tt.fields.saved,
				lastResponse:     tt.fields.lastResponse,
				lastResponseBody: tt.fields.lastResponseBody,
				isDebug:          tt.fields.isDebug,
			}
			if err := af.TheResponseShouldBeInXML(); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldBeInXML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldNotBe(t *testing.T) {
	type fields struct {
		saved            map[string]interface{}
		lastResponse     *http.Response
		lastResponseBody []byte
		isDebug          bool
	}
	type args struct {
		node   string
		goType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "is not nil value", fields: fields{
			lastResponseBody: []byte(`{
	"user": "abc"
}`),
		}, args: args{node: "user", goType: "nil"}, wantErr: false},
		{name: "is nil value", fields: fields{
			lastResponseBody: []byte(`{
	"user": nil
}`),
		}, args: args{node: "user", goType: "nil"}, wantErr: true},
		{name: "is null value", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "nil"}, wantErr: true},
		{name: "is not string #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "string"}, wantErr: false},
		{name: "is not string #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": 2
}`),
		}, args: args{node: "user", goType: "string"}, wantErr: false},
		{name: "is string", fields: fields{
			lastResponseBody: []byte(`{
	"user": "abc"
}`),
		}, args: args{node: "user", goType: "string"}, wantErr: true},
		{name: "is not int #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is not int #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": 2.1
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is int #1 <- special case", fields: fields{
			lastResponseBody: []byte(`{
	"user": 2.0
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is int #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": -1
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is float", fields: fields{
			lastResponseBody: []byte(`{
	"user": -1.0
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is not float #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": -1
}`),
		}, args: args{node: "user", goType: "float"}, wantErr: false},
		{name: "is not float #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "float"}, wantErr: false},
		{name: "is not float #3", fields: fields{
			lastResponseBody: []byte(`{
	"user": true
}`),
		}, args: args{node: "user", goType: "float"}, wantErr: false},
		{name: "is bool", fields: fields{
			lastResponseBody: []byte(`{
	"user": true
}`),
		}, args: args{node: "user", goType: "bool"}, wantErr: true},
		{name: "is not bool #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "bool"}, wantErr: false},
		{name: "is not bool #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": "false"
}`),
		}, args: args{node: "user", goType: "bool"}, wantErr: false},
		{name: "is map #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": {}
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is map #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": {"name": "pawel"}
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is not map #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is not map #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": "pawel"
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is slice #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": []
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "is slice #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": ["1"]
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "is not slice #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": "xxx"
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "is not slice #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": {}
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "unknown type", fields: fields{
			lastResponseBody: []byte(`{
	"user": {}
}`),
		}, args: args{node: "user", goType: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &ApiFeature{
				saved:            tt.fields.saved,
				lastResponse:     tt.fields.lastResponse,
				lastResponseBody: tt.fields.lastResponseBody,
				isDebug:          tt.fields.isDebug,
			}
			if err := af.TheJSONNodeShouldNotBe(tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldNotBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldBe(t *testing.T) {
	type fields struct {
		saved            map[string]interface{}
		lastResponse     *http.Response
		lastResponseBody []byte
		isDebug          bool
	}
	type args struct {
		node   string
		goType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "is not nil value", fields: fields{
			lastResponseBody: []byte(`{
	"user": "abc"
}`),
		}, args: args{node: "user", goType: "nil"}, wantErr: true},
		{name: "is nil value", fields: fields{
			lastResponseBody: []byte(`{
	"user": nil
}`),
		}, args: args{node: "user", goType: "nil"}, wantErr: false},
		{name: "is null value", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "nil"}, wantErr: false},
		{name: "is not string #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "string"}, wantErr: true},
		{name: "is not string #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": 2
}`),
		}, args: args{node: "user", goType: "string"}, wantErr: true},
		{name: "is string", fields: fields{
			lastResponseBody: []byte(`{
	"user": "abc"
}`),
		}, args: args{node: "user", goType: "string"}, wantErr: false},
		{name: "is not int #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is not int #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": 2.1
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is int #1 <- special case", fields: fields{
			lastResponseBody: []byte(`{
	"user": 2.0
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is int #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": -1
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is float", fields: fields{
			lastResponseBody: []byte(`{
	"user": -1.0
}`),
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is not float #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": -1
}`),
		}, args: args{node: "user", goType: "float"}, wantErr: true},
		{name: "is not float #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "float"}, wantErr: true},
		{name: "is not float #3", fields: fields{
			lastResponseBody: []byte(`{
	"user": true
}`),
		}, args: args{node: "user", goType: "float"}, wantErr: true},
		{name: "is bool", fields: fields{
			lastResponseBody: []byte(`{
	"user": true
}`),
		}, args: args{node: "user", goType: "bool"}, wantErr: false},
		{name: "is not bool #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "bool"}, wantErr: true},
		{name: "is not bool #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": "false"
}`),
		}, args: args{node: "user", goType: "bool"}, wantErr: true},
		{name: "is map #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": {}
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is map #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": {"name": "pawel"}
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is not map #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": null
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is not map #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": "pawel"
}`),
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is slice #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": []
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "is slice #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": ["1"]
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "is not slice #1", fields: fields{
			lastResponseBody: []byte(`{
	"user": "xxx"
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "is not slice #2", fields: fields{
			lastResponseBody: []byte(`{
	"user": {}
}`),
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "unknown type", fields: fields{
			lastResponseBody: []byte(`{
	"user": {}
}`),
		}, args: args{node: "user", goType: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &ApiFeature{
				saved:            tt.fields.saved,
				lastResponse:     tt.fields.lastResponse,
				lastResponseBody: tt.fields.lastResponseBody,
				isDebug:          tt.fields.isDebug,
			}
			if err := af.TheJSONNodeShouldBe(tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
