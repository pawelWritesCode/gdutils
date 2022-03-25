package schema

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

type mockedFileValidator struct {
	mock.Mock
}

type mockedUrlValidator struct {
	mock.Mock
}

func (m *mockedFileValidator) Validate(in any) error {
	args := m.Called(in)

	return args.Error(0)
}

func (m *mockedUrlValidator) Validate(in any) error {
	args := m.Called(in)

	return args.Error(0)
}

func TestJSONSchemaValidator_getSource(t *testing.T) {
	type fields struct {
		fileValidator validator.Validator
		urlValidator  validator.Validator
		schemasDir    string
		mockFunc      func()
	}
	type args struct {
		rawSource string
	}

	mFileValidator := new(mockedFileValidator)
	mUrlValidator := new(mockedUrlValidator)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "is empty string", fields: fields{
			fileValidator: mFileValidator,
			urlValidator:  mUrlValidator,
			schemasDir:    "",
			mockFunc: func() {

			},
		}, args: args{rawSource: ""}, want: "", wantErr: true},
		{name: "is not valid URl and is not valid path", fields: fields{
			fileValidator: mFileValidator,
			urlValidator:  mUrlValidator,
			schemasDir:    "",
			mockFunc: func() {
				mUrlValidator.On("Validate", "/json_schema").Return(errors.New("a")).Once()
				mFileValidator.On("Validate", "/json_schema").Return(errors.New("b")).Once()
			},
		}, args: args{rawSource: "/json_schema"}, want: "", wantErr: true},
		{name: "is valid URL", fields: fields{
			fileValidator: mFileValidator,
			urlValidator:  mUrlValidator,
			schemasDir:    "",
			mockFunc: func() {
				mUrlValidator.On("Validate", "www.json-schema.org/user.json").Return(nil).Once()
				mFileValidator.On("Validate", "www.json-schema.org/user.json").Return(errors.New("b")).Once()
			},
		}, args: args{rawSource: "www.json-schema.org/user.json"}, want: "www.json-schema.org/user.json", wantErr: false},
		{name: "is valid path on user OS", fields: fields{
			fileValidator: mFileValidator,
			urlValidator:  mUrlValidator,
			schemasDir:    "",
			mockFunc: func() {
				mUrlValidator.On("Validate", "/jsons/user.json").Return(errors.New("a")).Once()
				mFileValidator.On("Validate", "/jsons/user.json").Return(nil).Once()
			},
		}, args: args{rawSource: "/jsons/user.json"}, want: "file:///jsons/user.json", wantErr: false},
		{name: "is valid path on user OS", fields: fields{
			fileValidator: mFileValidator,
			urlValidator:  mUrlValidator,
			schemasDir:    "/jsons",
			mockFunc: func() {
				mUrlValidator.On("Validate", "user.json").Return(errors.New("a")).Once()
				mFileValidator.On("Validate", "/jsons/user.json").Return(nil).Once()
			},
		}, args: args{rawSource: "user.json"}, want: "file:///jsons/user.json", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsv := JSONSchemaReferenceValidator{
				fileValidator: tt.fields.fileValidator,
				urlValidator:  tt.fields.urlValidator,
				schemasDir:    tt.fields.schemasDir,
			}

			tt.fields.mockFunc()

			got, err := jsv.getSource(tt.args.rawSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSource() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONSchemaRawValidator_Validate(t *testing.T) {
	document := `{
  "latitude": 48.858093,
  "longitude": 2.294694
}`
	jsonSchema := `{
  "$id": "https://example.com/geographical-location.schema.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Longitude and Latitude Values",
  "description": "A geographical coordinate.",
  "required": [ "latitude", "longitude" ],
  "type": "object",
  "properties": {
    "latitude": {
      "type": "number",
      "minimum": -90,
      "maximum": 90
    },
    "longitude": {
      "type": "number",
      "minimum": -180,
      "maximum": 180
    }
  }
}`

	type args struct {
		document   string
		jsonSchema string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "valid data #1", args: args{
			document:   document,
			jsonSchema: jsonSchema,
		}, wantErr: false},
		{name: "no document", args: args{
			document:   "",
			jsonSchema: jsonSchema,
		}, wantErr: true},
		{name: "no json schema", args: args{
			document:   document,
			jsonSchema: ``,
		}, wantErr: true},
		{name: "invalid json schema", args: args{
			document: document,
			jsonSchema: `{
  "$id": "https://example.com/geographical-location.schema.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Longitude and Latitude Values",
  "description": "A geographical coordinate.",
  "required": [ "latitude", "longitude" ],
  "type": "object",
  "properties": {
    "latitude": {
      "type": "string"
    },
    "longitude": {
      "type": "string"
    }
  }
}`,
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			J := JSONSchemaRawValidator{}
			if err := J.Validate(tt.args.document, tt.args.jsonSchema); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
