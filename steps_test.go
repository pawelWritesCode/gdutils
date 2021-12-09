package gdutils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/mock"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/template"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

type mockedHTTPContext struct {
	mock.Mock
}

type mockedJSONValidator struct {
	mock.Mock
}

func (m mockedJSONValidator) Validate(document, schemaPath string) error {
	args := m.Called(document, schemaPath)

	return args.Error(0)
}

func (m mockedHTTPContext) GetHTTPClient() *http.Client {
	args := m.Called()

	return args.Get(0).(*http.Client)
}

func (m mockedHTTPContext) GetLastResponse() (*http.Response, error) {
	args := m.Called()

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m mockedHTTPContext) GetLastResponseBody() ([]byte, error) {
	args := m.Called()

	return args.Get(0).([]byte), args.Error(1)
}

func TestApiFeature_theJSONNodeShouldBeOfValue(t *testing.T) {
	type fields struct {
		lastResponse *http.Response
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
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(``))},
		}, args: args{
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: true},
		{name: "json with first level field with string data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"name": "ivo"
}`))},
		}, args: args{
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: false},
		{name: "json with first level field with int data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10
}`))},
		}, args: args{
			expr:      "number",
			dataType:  "int",
			dataValue: "10",
		}, wantErr: false},
		{name: "json with first level field with float64 data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10.1
}`))},
		}, args: args{
			expr:      "number",
			dataType:  "float",
			dataValue: "10.1",
		}, wantErr: false},
		{name: "json with first level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"is": true
}`))},
		}, args: args{
			expr:      "is",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data": {
		"name": "Is empty",
		"value": true
	}
}`))},
		}, args: args{
			expr:      "data.value",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
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
		}, args: args{
			expr:      "data[1].value",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data":	[
			true,
			false
		]
}`))},
		}, args: args{
			expr:      "data[1]",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultState(false, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.TheJSONNodeShouldBeOfValue(tt.args.expr, tt.args.dataType, tt.args.dataValue); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiFeature_TheJSONNodeShouldBeSliceOfLength(t *testing.T) {
	type fields struct {
		lastResponse *http.Response
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
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))},
		}, args: args{
			expr:   "anykey",
			length: 0,
		}, wantErr: true},
		{name: "key is not slice", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"name": "xyz"	
}`))},
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"name": {
		"details": "xyz"
	}
}`))},
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is slice but length does not match", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"names": ["a", "b"]
}`))},
		}, args: args{
			expr:   "name",
			length: 0,
		}, wantErr: true},
		{name: "key is slice and length match", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"names": ["a", "b"]
}`))},
		}, args: args{
			expr:   "names",
			length: 2,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultState(false, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)
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
			af := NewDefaultState(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

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
			af := NewDefaultState(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

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
		code int
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
			s := NewDefaultState(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := s.TheResponseStatusCodeShouldBe(tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseStatusCodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScenario_ISaveFromTheLastResponseJSONNodeAs(t *testing.T) {
	type fields struct {
		cache        cache.Cache
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
			cache: cache.New(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{node: "token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.New(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{node: "last_name", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.New(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{node: "user.last_name", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.New(),
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
			s := NewDefaultState(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			err := s.ISaveFromTheLastResponseJSONNodeAs(tt.args.node, tt.args.variableName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ISaveFromTheLastResponseJSONNodeAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if _, err := s.Cache.GetSaved(tt.args.variableName); err != nil {
					t.Errorf("%s was not saved to Cache", tt.args.node)
				}
			}
		})
	}
}

func TestScenario_IGenerateARandomIntInTheRangeToAndSaveItAs(t *testing.T) {
	s := NewDefaultState(false, "")
	for i := 0; i < 100; i++ {
		if err := s.IGenerateARandomIntInTheRangeToAndSaveItAs(0, 100000, "RANDOM_INT"); (err != nil) != false {
			t.Errorf("IGenerateARandomIntInTheRangeToAndSaveItAs() error = %v, wantErr %v", err, false)
		}

		randomInteger, err := s.Cache.GetSaved("RANDOM_INT")
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
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: ""}, wantErr: true},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type"}, wantErr: false},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "Content-Type"}, wantErr: false},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": {"30"},
				"Content-Type":   {"application/json"},
			},
		},
		}, args: args{name: "Content-Type"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultState(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

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
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "", value: "application/json"}, wantErr: true},
		{name: "matching header but improper value", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type", value: "application/xml"}, wantErr: true},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type", value: "application/json"}, wantErr: false},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "Content-Type", value: "application/json"}, wantErr: false},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": {"30"},
				"Content-Type":   {"application/json"},
			},
		},
		}, args: args{name: "Content-Type", value: "application/json"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultState(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := s.TheResponseShouldHaveHeaderOfValue(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveHeaderOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_IPrepareNewRequestToAndSaveItAs(t *testing.T) {
	type fields struct {
		IsDebug bool
	}
	type args struct {
		method      string
		urlTemplate string
		cacheKey    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			fields:  fields{IsDebug: false},
			args:    args{method: http.MethodGet, urlTemplate: "/", cacheKey: "MY_GET_REQUEST"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultState(tt.fields.IsDebug, "")
			if err := s.IPrepareNewRequestToAndSaveItAs(tt.args.method, tt.args.urlTemplate, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("IPrepareNewRequestToAndSaveItAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				req, err := s.getPreparedRequest(tt.args.cacheKey)
				if err != nil {
					t.Errorf("%v", err)
				}

				if req.Method != tt.args.method {
					t.Errorf("obtained request has different method: %s, expected: %s", req.Method, tt.args.method)
				}
			}
		})
	}
}

func TestState_ISetFollowingHeadersForPreparedRequest(t *testing.T) {
	type fields struct {
		IsDebug   bool
		reqMethod string
		reqUri    string
		cacheKey  string
	}
	type args struct {
		cacheKey        string
		headersTemplate *godog.DocString
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "invalid headers",
			fields:  fields{IsDebug: false},
			args:    args{cacheKey: "", headersTemplate: &godog.DocString{Content: "abc"}},
			wantErr: true,
		},
		{
			name:    "no request",
			fields:  fields{IsDebug: false},
			args:    args{cacheKey: "abc", headersTemplate: &godog.DocString{Content: `{"Content-Type": "application/json"}`}},
			wantErr: true,
		},
		{
			name:    "cache key does not point at request",
			fields:  fields{IsDebug: false, reqMethod: "GET", reqUri: "/", cacheKey: "abc"},
			args:    args{cacheKey: "xxx", headersTemplate: &godog.DocString{Content: `{"Content-Type": "application/json"}`}},
			wantErr: true,
		},
		{
			name:    "successfully set request header",
			fields:  fields{IsDebug: false, reqMethod: "GET", reqUri: "/", cacheKey: "abc"},
			args:    args{cacheKey: "abc", headersTemplate: &godog.DocString{Content: `{"Content-Type": "application/json"}`}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultState(tt.fields.IsDebug, "")

			err := s.IPrepareNewRequestToAndSaveItAs(tt.fields.reqMethod, tt.fields.reqUri, tt.fields.cacheKey)
			if err != nil {
				t.Errorf("%v", err)
			}

			if err = s.ISetFollowingHeadersForPreparedRequest(tt.args.cacheKey, tt.args.headersTemplate); (err != nil) != tt.wantErr {
				t.Errorf("ISetFollowingHeadersForPreparedRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_ISetFollowingBodyForPreparedRequest(t *testing.T) {
	type fields struct {
		IsDebug   bool
		reqMethod string
		reqUri    string
		cacheKey  string
	}
	type args struct {
		cacheKey     string
		bodyTemplate *godog.DocString
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "no request",
			fields:  fields{IsDebug: false},
			args:    args{cacheKey: "abc", bodyTemplate: &godog.DocString{Content: `{"Content-Type": "application/json"}`}},
			wantErr: true,
		},
		{
			name:    "cache key does not point at request",
			fields:  fields{IsDebug: false, reqMethod: "GET", reqUri: "/", cacheKey: "abc"},
			args:    args{cacheKey: "xxx", bodyTemplate: &godog.DocString{Content: `{"Content-Type": "application/json"}`}},
			wantErr: true,
		},
		{
			name:    "successfully set request body",
			fields:  fields{IsDebug: false, reqMethod: "GET", reqUri: "/", cacheKey: "abc"},
			args:    args{cacheKey: "abc", bodyTemplate: &godog.DocString{Content: `{"a": "b"}`}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultState(tt.fields.IsDebug, "")
			err := s.IPrepareNewRequestToAndSaveItAs(tt.fields.reqMethod, tt.fields.reqUri, tt.fields.cacheKey)
			if err != nil {
				t.Errorf("%v", err)
			}

			if err := s.ISetFollowingBodyForPreparedRequest(tt.args.cacheKey, tt.args.bodyTemplate); (err != nil) != tt.wantErr {
				t.Errorf("ISetFollowingBodyForPreparedRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_IValidateLastResponseBodyWithSchema(t *testing.T) {
	type fields struct {
		Debugger            debugger.Debugger
		HttpContext         httpctx.HttpContext
		TemplateEngine      template.TemplateEngine
		JSONSchemaValidator validator.SchemaValidator
		mockFunc            func()
	}
	type args struct {
		schemaPath string
	}

	mHttpContext := new(mockedHTTPContext)
	mJSONValidator := new(mockedJSONValidator)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "missing last response body", fields: fields{
			Debugger:            nil,
			TemplateEngine:      nil,
			JSONSchemaValidator: mJSONValidator,
			HttpContext:         mHttpContext,
			mockFunc: func() {
				mHttpContext.On("GetLastResponseBody").Return([]byte(""), fmt.Errorf("abc")).Once()
			},
		}, args: args{schemaPath: ""}, wantErr: true},
		{name: "validator fails", fields: fields{
			Debugger:            nil,
			TemplateEngine:      nil,
			JSONSchemaValidator: mJSONValidator,
			HttpContext:         mHttpContext,
			mockFunc: func() {
				mHttpContext.On("GetLastResponseBody").Return([]byte(""), nil).Once()
				mJSONValidator.On("Validate", "", "").Return(errors.New("abc")).Once()
			},
		}, args: args{schemaPath: ""}, wantErr: true},
		{name: "validator succed", fields: fields{
			Debugger:            nil,
			TemplateEngine:      nil,
			JSONSchemaValidator: mJSONValidator,
			HttpContext:         mHttpContext,
			mockFunc: func() {
				mHttpContext.On("GetLastResponseBody").Return([]byte(""), nil).Once()
				mJSONValidator.On("Validate", "", "").Return(nil).Once()
			},
		}, args: args{schemaPath: ""}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				Debugger:            tt.fields.Debugger,
				Cache:               cache.New(),
				HttpContext:         tt.fields.HttpContext,
				TemplateEngine:      tt.fields.TemplateEngine,
				JSONSchemaValidator: tt.fields.JSONSchemaValidator,
			}

			tt.fields.mockFunc()
			if err := s.IValidateLastResponseBodyWithSchema(tt.args.schemaPath); (err != nil) != tt.wantErr {
				t.Errorf("IValidateLastResponseBodyWithSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
