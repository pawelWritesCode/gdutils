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
				baseUrl:          tt.fields.baseUrl,
			}
			if err := af.TheJSONNodeShouldBeOfValue(tt.args.expr, tt.args.dataType, tt.args.dataValue); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
