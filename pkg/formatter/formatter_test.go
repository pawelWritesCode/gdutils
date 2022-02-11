package formatter

import (
	"testing"
)

type bodyHeaders struct {
	// Body should contain HTTP(s) request body
	Body interface{} `xml:"body" yaml:"body"`

	// Headers should contain HTTP(s) request headers
	Headers map[string]string `xml:"headers" yaml:"headers"`
}

func TestJSONFormatter_Deserialize(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "no data", args: args{
			data: nil,
		}, wantErr: true},
		{name: "invalid data", args: args{
			data: []byte("abc"),
		}, wantErr: true},
		{name: "different format data", args: args{
			data: []byte("<body>abc</body>"),
		}, wantErr: true},
		{name: "proper data #1", args: args{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := bodyHeaders{}
			J := JSONFormatter{}
			if err := J.Deserialize(tt.args.data, &v); (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLFormatter_Deserialize(t *testing.T) {
	type args struct {
		data []byte
		v    interface{}
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