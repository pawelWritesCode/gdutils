package osutils

import (
	"fmt"
	"os"
	"strings"

	v "github.com/pawelWritesCode/gdutils/pkg/validator"
)

// FileValidator has ability to validate whether path points at any file on user OS
type FileValidator struct{}

// OSFileRecognizer is entity that has ability to recognize reference to file in OS from string
type OSFileRecognizer struct {
	fileValidator v.Validator

	prefix string
}

// FileReference describes found reference to file
type FileReference struct {
	// FoundPrefix holds information about found prefix
	FoundPrefix FoundPrefix

	// Reference holds information about found file reference
	Reference Reference
}

// FoundPrefix describes found prefix in process of file reference recognize
type FoundPrefix struct {
	// Index is byte index of first occurrence of prefix
	Index int

	// Value is prefix name
	Value string
}

// Reference holds information about resource reference
type Reference struct {
	// Value is raw value of reference
	Value string

	// Type is reference type
	Type ReferenceType
}

// ReferenceType describes type of reference
type ReferenceType string

const (
	// ReferenceTypeOSPath describes operating system path
	ReferenceTypeOSPath ReferenceType = "OS_PATH"
)

func NewFileValidator() FileValidator {
	return FileValidator{}
}

// NewOSFileRecognizer returns ready to work OSFileRecognizer. prefix should be fixed prefix of file
func NewOSFileRecognizer(prefix string, fileValidator v.Validator) OSFileRecognizer {
	return OSFileRecognizer{prefix: prefix, fileValidator: fileValidator}
}

// Validate checks whether in is valid path to any file on local user OS
func (fv FileValidator) Validate(in any) error {
	p, ok := in.(string)
	if !ok {
		return fmt.Errorf("%+v is not string", in)
	}

	_, err := os.Stat(p)

	isNotExist := os.IsNotExist(err)
	if isNotExist {
		return fmt.Errorf("%s does not point at any file in your local OS", p)
	}

	return nil
}

// Recognize accepts any string and look after reference to file as defined during construction phase in prefix field.
// second bool argument tells whether reference was found and is valid
func (fr OSFileRecognizer) Recognize(input string) (FileReference, bool) {
	fileReference := FileReference{}
	idx := strings.Index(input, fr.prefix)

	isFound := idx != -1

	if isFound {
		fileReference.FoundPrefix.Value = fr.prefix
		fileReference.FoundPrefix.Index = idx

		ref := input[idx+len(fr.prefix):]

		fileErr := fr.fileValidator.Validate(ref)
		if fileErr != nil {
			return fileReference, false
		}

		fileReference.Reference.Type = ReferenceTypeOSPath
		fileReference.Reference.Value = ref
	}

	return fileReference, isFound
}

// IsFoundReference tells whether reference to file was detected. It may not be valid
func (fr FileReference) IsFoundReference() bool {
	return len(fr.FoundPrefix.Value) > 0
}
