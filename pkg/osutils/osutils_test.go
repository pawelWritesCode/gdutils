package osutils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

type mockedFileValidator struct {
	mock.Mock
}

func (m *mockedFileValidator) Validate(in interface{}) error {
	args := m.Called(in)

	return args.Error(0)
}

func TestFileRecognizer_Recognize(t *testing.T) {
	mFileValidator := new(mockedFileValidator)

	type fields struct {
		fileValidator validator.Validator
		prefix        string
		mockFunc      func()
	}
	type args struct {
		input string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   FileReference
		want1  bool
	}{
		{name: "input does not contain file reference", fields: fields{
			fileValidator: mFileValidator,
			prefix:        "file://",
			mockFunc: func() {

			},
		}, args: args{input: ""}, want: FileReference{}, want1: false},
		{name: "input contains file reference, but reference is not valid", fields: fields{
			fileValidator: mFileValidator,
			prefix:        "file://",
			mockFunc: func() {
				mFileValidator.On("Validate", "abc").Return(errors.New("error")).Once()
			},
		}, args: args{input: "file://abc"}, want: FileReference{
			FoundPrefix: FoundPrefix{
				Value: "file://",
			},
		}, want1: false},

		{name: "input contains file reference #1", fields: fields{
			fileValidator: mFileValidator,
			prefix:        "file://",
			mockFunc: func() {
				mFileValidator.On("Validate", "/usr/local/bin/ls").Return(nil).Once()
			},
		}, args: args{input: "file:///usr/local/bin/ls"}, want: FileReference{
			FoundPrefix: FoundPrefix{
				Index: 0,
				Value: "file://",
			},
			Reference: Reference{
				Value: "/usr/local/bin/ls",
				Type:  ReferenceTypeOSPath,
			},
		}, want1: true},
		{name: "input contains file reference #2", fields: fields{
			fileValidator: mFileValidator,
			prefix:        "file://",
			mockFunc: func() {
				mFileValidator.On("Validate", "/usr/local/bin/ls").Return(nil).Once()
			},
		}, args: args{input: "abc file:///usr/local/bin/ls"}, want: FileReference{
			FoundPrefix: FoundPrefix{
				Index: 4,
				Value: "file://",
			},
			Reference: Reference{
				Value: "/usr/local/bin/ls",
				Type:  ReferenceTypeOSPath,
			},
		}, want1: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := OSFileRecognizer{
				fileValidator: tt.fields.fileValidator,
				prefix:        tt.fields.prefix,
			}

			tt.fields.mockFunc()

			got, got1 := fr.Recognize(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Recognize() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Recognize() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
