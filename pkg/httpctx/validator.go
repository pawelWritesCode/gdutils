package httpctx

import (
	"fmt"
	"net/url"
)

// URLValidator has ability to validate URL
type URLValidator struct{}

func NewURLValidator() URLValidator { return URLValidator{} }

// Validate checks whether in is valid URL
func (U URLValidator) Validate(in interface{}) error {
	source, ok := in.(string)
	if !ok {
		return fmt.Errorf("+%v is not string", in)
	}

	u, err := url.ParseRequestURI(source)
	if err == nil {
		if u.Scheme != "" && u.Host != "" {
			return nil
		}
	}

	return fmt.Errorf("%s is invalid URL", source)
}
