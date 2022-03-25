package formatter

import (
	"net/http"
	"testing"
)

type bodyHeaders struct {
	// Body should contain HTTP(s) request body
	Body any `xml:"body" yaml:"body"`

	// Headers should contain HTTP(s) request headers
	Headers map[string]string `xml:"headers" yaml:"headers"`
}

type cookiesSlice []http.Cookie

func TestJSONFormatter_Deserialize(t *testing.T) {

	type fields struct {
		v any
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "no data", fields: fields{v: bodyHeaders{}}, args: args{
			data: nil,
		}, wantErr: true},
		{name: "invalid data", fields: fields{v: bodyHeaders{}}, args: args{
			data: []byte("abc"),
		}, wantErr: true},
		{name: "different format data", fields: fields{v: bodyHeaders{}}, args: args{
			data: []byte("<body>abc</body>"),
		}, wantErr: true},
		{name: "proper data #1", fields: fields{v: bodyHeaders{}}, args: args{
			data: []byte(`{
        "body": {
            "firstName": "{{.RANDOM_FIRST_NAME}}",
            "lastName": "{{.RANDOM_LAST_NAME}}",
            "age": 12
        },
        "headers": {
            "Content-Type": "application/json"
        }
    }`),
		}, wantErr: false},
		{name: "proper data #1", fields: fields{v: cookiesSlice{}}, args: args{
			data: []byte(`[
		{
			"name": "token",
			"value": "abc",
			"expires": "{{.NOW.Format \"Jan 2, 2006 at 3:04pm (MST)\"}}"
		}
		]`),
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//v := bodyHeaders{}
			J := JSONFormatter{}
			if err := J.Deserialize(tt.args.data, &tt.fields.v); (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLFormatter_Deserialize(t *testing.T) {
	type args struct {
		data []byte
		v    any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no data #1", args: args{
			data: nil,
		}, wantErr: true},
		{name: "no data #2", args: args{
			data: []byte(""),
		}, wantErr: true},
		{name: "invalid data", args: args{
			data: []byte("abc"),
		}, wantErr: true},
		{name: "different format data", args: args{
			data: []byte(`{
       "body": {
           "firstName": "{{.RANDOM_FIRST_NAME}}",
           "lastName": "{{.RANDOM_LAST_NAME}}",
           "age": 12
       },
       "headers": {
           "Content-Type": "application/json"
       }
    }`),
		}, wantErr: true},
		{name: "proper data format", args: args{
			data: []byte(`---
body:
  firstName: "{{.RANDOM_FIRST_NAME}}"
  lastName: "{{.RANDOM_LAST_NAME}}"
  age: 12
headers:
  Content-Type: application/json`),
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := bodyHeaders{}

			Y := YAMLFormatter{}
			if err := Y.Deserialize(tt.args.data, &v); (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAwareFormatter_Deserialize(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no data #1", args: args{
			data: nil,
		}, wantErr: true},
		{name: "no data #2", args: args{
			data: []byte(""),
		}, wantErr: true},
		{name: "invalid data", args: args{
			data: []byte("abc"),
		}, wantErr: true},
		{name: "proper data format #1", args: args{
			data: []byte(`{
        "body": {
            "firstName": "{{.RANDOM_FIRST_NAME}}",
            "lastName": "{{.RANDOM_LAST_NAME}}",
            "age": 12
        },
        "headers": {
            "Content-Type": "application/json"
        }
    }`),
		}, wantErr: false},
		{name: "proper data format #2", args: args{
			data: []byte(`---
body:
  firstName: "{{.RANDOM_FIRST_NAME}}"
  lastName: "{{.RANDOM_LAST_NAME}}"
  age: 12
headers:
  Content-Type: application/json`),
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := bodyHeaders{}

			a := AwareFormatter{}
			if err := a.Deserialize(tt.args.data, &v); (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestXMLFormatter_Deserialize(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no data #1", args: args{
			data: nil,
		}, wantErr: true},
		{name: "no data #2", args: args{
			data: []byte(""),
		}, wantErr: true},
		{name: "invalid data", args: args{
			data: []byte("abc"),
		}, wantErr: true},
		{name: "proper data format #1", args: args{
			data: []byte(`{
        "body": {
            "firstName": "{{.RANDOM_FIRST_NAME}}",
            "lastName": "{{.RANDOM_LAST_NAME}}",
            "age": 12
        },
        "headers": {
            "Content-Type": "application/json"
        }
    }`),
		}, wantErr: true},
		{name: "proper data format #2", args: args{
			data: []byte(`---
body:
  firstName: "{{.RANDOM_FIRST_NAME}}"
  lastName: "{{.RANDOM_LAST_NAME}}"
  age: 12
headers:
  Content-Type: application/json`),
		}, wantErr: true},

		{name: "XML format data #1", args: args{
			data: []byte("<body>abc</body>"),
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			X := XMLFormatter{}
			v := bodyHeaders{}

			if err := X.Deserialize(tt.args.data, &v); (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
