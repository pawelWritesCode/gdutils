package osutils

import (
	"fmt"
	"os"
)

// FileValidator has ability to validate whether path points at any file on user OS
type FileValidator struct{}

func NewFileValidator() FileValidator {
	return FileValidator{}
}

// Validate checks whether in is valid path to any file on local user OS
func (fv FileValidator) Validate(in interface{}) error {
	p, ok := in.(string)
	if !ok {
		return fmt.Errorf("%+v is not string", in)
	}

	_, err := os.Stat(p)

	isNotExist := os.IsNotExist(err)
	if !isNotExist {
		return nil
	}

	return fmt.Errorf("%s does not point at any file in your local OS", p)
}
