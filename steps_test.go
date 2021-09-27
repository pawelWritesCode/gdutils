package gdutils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestApiFeature_theJSONNodeShouldBeOfValue(t *testing.T) {
	type fields struct {
		saved        map[string]interface{}
		lastResponse *http.Response
		baseUrl      string
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
			saved:        nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(``))},
			baseUrl:      "",
		}, args: args{
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: true},
		{name: "json with first level field with string data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"name": "ivo"
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: false},
		{name: "json with first level field with int data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "number",
			dataType:  "int",
			dataValue: "10",
		}, wantErr: false},
		{name: "json with first level field with float64 data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10.1
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "number",
			dataType:  "float",
			dataValue: "10.1",
		}, wantErr: false},
		{name: "json with first level field with bool data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"is": true
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "is",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data": {
		"name": "Is empty",
		"value": true
	}
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "data.value",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
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
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "data[1].value",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data":	[
			true,
			false
		]
}`))},
			baseUrl: "",
		}, args: args{
			expr:      "data[1]",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &State{
				cache:        tt.fields.saved,
				lastResponse: tt.fields.lastResponse,
			}
			if err := af.TheJSONNodeShouldBeOfValue(tt.args.expr, tt.args.dataType, tt.args.dataValue); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldBeSliceOfLength(t *testing.T) {
	type fields struct {
		saved        map[string]interface{}
		lastResponse *http.Response
		baseUrl      string
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
			saved:        nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))},
			baseUrl:      "",
		}, args: args{
			expr:   "anykey",
			length: 0,
		}, wantErr: true},
		{name: "key is not slice", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"name": "xyz"	
}`))},
			baseUrl: "",
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is not slice #2", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"name": {
		"details": "xyz"
	}
}`))},
			baseUrl: "",
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is slice but length does not match", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"names": ["a", "b"]
}`))},
			baseUrl: "",
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is slice and length match", fields: fields{
			saved: nil,
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"names": ["a", "b"]
}`))},
			baseUrl: "",
		}, args: args{
			expr:   "names",
			length: 2,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &State{
				cache:        tt.fields.saved,
				lastResponse: tt.fields.lastResponse,
			}
			if err := af.TheJSONNodeShouldBeSliceOfLength(tt.args.expr, tt.args.length); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeSliceOfLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldNotBe(t *testing.T) {
	type fields struct {
		saved        map[string]interface{}
		lastResponse *http.Response
		isDebug      bool
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
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{node: "user", goType: "nil"}, wantErr: false},
		{name: "is nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": nil
}`))},
		}, args: args{node: "user", goType: "nil"}, wantErr: true},
		{name: "is null value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "nil"}, wantErr: true},
		{name: "is not string #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "string"}, wantErr: false},
		{name: "is not string #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2
}`))},
		}, args: args{node: "user", goType: "string"}, wantErr: false},
		{name: "is string", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{node: "user", goType: "string"}, wantErr: true},
		{name: "is not int #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is not int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.1
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is int #1 <- special case", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.0
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is float", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1.0
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is not float #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{node: "user", goType: "float"}, wantErr: false},
		{name: "is not float #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "float"}, wantErr: false},
		{name: "is not float #3", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{node: "user", goType: "float"}, wantErr: false},
		{name: "is bool", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{node: "user", goType: "bool"}, wantErr: true},
		{name: "is not bool #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "bool"}, wantErr: false},
		{name: "is not bool #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "false"
}`))},
		}, args: args{node: "user", goType: "bool"}, wantErr: false},
		{name: "is map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {"name": "pawel"}
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is not map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is not map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "pawel"
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": []
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "is slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": ["1"]
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "is not slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "xxx"
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "unknown type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &State{
				cache:        tt.fields.saved,
				lastResponse: tt.fields.lastResponse,
				isDebug:      tt.fields.isDebug,
			}
			if err := af.TheJSONNodeShouldNotBe(tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldNotBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldBe(t *testing.T) {
	type fields struct {
		saved        map[string]interface{}
		lastResponse *http.Response
		isDebug      bool
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
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{node: "user", goType: "nil"}, wantErr: true},
		{name: "is nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": nil
}`))},
		}, args: args{node: "user", goType: "nil"}, wantErr: false},
		{name: "is null value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "nil"}, wantErr: false},
		{name: "is not string #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "string"}, wantErr: true},
		{name: "is not string #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2
}`))},
		}, args: args{node: "user", goType: "string"}, wantErr: true},
		{name: "is string", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{node: "user", goType: "string"}, wantErr: false},
		{name: "is not int #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is not int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.1
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: true},
		{name: "is int #1 <- special case", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.0
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is float", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1.0
}`))},
		}, args: args{node: "user", goType: "int"}, wantErr: false},
		{name: "is not float #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{node: "user", goType: "float"}, wantErr: true},
		{name: "is not float #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "float"}, wantErr: true},
		{name: "is not float #3", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{node: "user", goType: "float"}, wantErr: true},
		{name: "is bool", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{node: "user", goType: "bool"}, wantErr: false},
		{name: "is not bool #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "bool"}, wantErr: true},
		{name: "is not bool #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "false"
}`))},
		}, args: args{node: "user", goType: "bool"}, wantErr: true},
		{name: "is map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {"name": "pawel"}
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: false},
		{name: "is not map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is not map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "pawel"
}`))},
		}, args: args{node: "user", goType: "map"}, wantErr: true},
		{name: "is slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": []
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "is slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": ["1"]
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: false},
		{name: "is not slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "xxx"
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "slice"}, wantErr: true},
		{name: "unknown type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &State{
				cache:        tt.fields.saved,
				lastResponse: tt.fields.lastResponse,
				isDebug:      tt.fields.isDebug,
			}
			if err := af.TheJSONNodeShouldBe(tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScenario_TheResponseStatusCodeShouldBe(t *testing.T) {
	type fields struct {
		cache        map[string]interface{}
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		code uint16
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "invalid code #1, code less than 200", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 1}, wantErr: true},
		{name: "invalid code #2, code over 599", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 600}, wantErr: true},
		{name: "invalid code #3", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 400}, wantErr: true},
		{name: "valid code #1", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 200}, wantErr: false},
		{name: "valid code #1", fields: fields{lastResponse: &http.Response{StatusCode: 400}}, args: args{code: 400}, wantErr: false},
		{name: "valid code #1", fields: fields{lastResponse: &http.Response{StatusCode: 511}}, args: args{code: 511}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				cache:        tt.fields.cache,
				lastResponse: tt.fields.lastResponse,
				isDebug:      tt.fields.isDebug,
			}
			if err := s.TheResponseStatusCodeShouldBe(tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseStatusCodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScenario_ISaveFromTheLastResponseJSONNodeAs(t *testing.T) {
	type fields struct {
		cache        map[string]interface{}
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		node         string
		variableName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "invalid node #1", fields: fields{
			cache: map[string]interface{}{},
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{node: "token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: map[string]interface{}{},
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{node: "last_name", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: map[string]interface{}{},
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{node: "user.last_name", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: map[string]interface{}{},
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{node: "user", variableName: "USER"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				cache:        tt.fields.cache,
				lastResponse: tt.fields.lastResponse,
				isDebug:      tt.fields.isDebug,
			}

			err := s.ISaveFromTheLastResponseJSONNodeAs(tt.args.node, tt.args.variableName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ISaveFromTheLastResponseJSONNodeAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if _, err := s.GetSaved(tt.args.variableName); err != nil {
					t.Errorf("%s was not saved to cache", tt.args.node)
				}
			}
		})
	}
}

func TestScenario_IGenerateARandomIntInTheRangeToAndSaveItAs(t *testing.T) {
	s := &State{
		cache:        map[string]interface{}{},
		lastResponse: nil,
		isDebug:      false,
	}
	for i := 0; i < 100; i++ {
		if err := s.IGenerateARandomIntInTheRangeToAndSaveItAs(0, 100000, "RANDOM_INT"); (err != nil) != false {
			t.Errorf("IGenerateARandomIntInTheRangeToAndSaveItAs() error = %v, wantErr %v", err, false)
		}

		randomInteger, err := s.GetSaved("RANDOM_INT")
		if err != nil {
			t.Errorf("%v", err)
		}

		randomInt := randomInteger.(int)
		if randomInt < 0 {
			t.Errorf("randomInt should not be less than 0")
		}

		if randomInt > 100000 {
			t.Errorf("randomInt should not be greater than 100000")
		}
	}
}

func TestScenario_TheResponseShouldHaveHeader(t *testing.T) {
	type fields struct {
		cache        map[string]interface{}
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "no headers in request", fields: fields{lastResponse: &http.Response{Header: map[string][]string{}}}, args: args{name: "Content-Type"}, wantErr: true},
		{name: "empty string provided as header name", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: ""}, wantErr: true},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: "content-type"}, wantErr: false},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: "Content-Type"}, wantErr: false},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": []string{"30"},
				"Content-Type":   []string{"application/json"},
			},
		},
		}, args: args{name: "Content-Type"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				cache:        tt.fields.cache,
				lastResponse: tt.fields.lastResponse,
				isDebug:      tt.fields.isDebug,
			}
			if err := s.TheResponseShouldHaveHeader(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScenario_TheResponseShouldHaveHeaderOfValue(t *testing.T) {
	type fields struct {
		cache        map[string]interface{}
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "no headers in request", fields: fields{lastResponse: &http.Response{Header: map[string][]string{}}}, args: args{name: "Content-Type", value: "application/json"}, wantErr: true},
		{name: "empty string provided as header name", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: "", value: "application/json"}, wantErr: true},
		{name: "matching header but improper value", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: "content-type", value: "application/xml"}, wantErr: true},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: "content-type", value: "application/json"}, wantErr: false},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": []string{"application/json"}}},
		}, args: args{name: "Content-Type", value: "application/json"}, wantErr: false},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": []string{"30"},
				"Content-Type":   []string{"application/json"},
			},
		},
		}, args: args{name: "Content-Type", value: "application/json"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				cache:        tt.fields.cache,
				lastResponse: tt.fields.lastResponse,
				isDebug:      tt.fields.isDebug,
			}
			if err := s.TheResponseShouldHaveHeaderOfValue(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveHeaderOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
