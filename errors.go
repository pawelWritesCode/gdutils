package gdutils

import (
	"errors"
	"fmt"
)

//ErrGdutils is general package error and can be tested against
var ErrGdutils = errors.New("gdutils")

//ErrJson tells that there is problem with JSON
var ErrJson = fmt.Errorf("%w: something wrong with JSON", ErrGdutils)

//ErrHTTPReqRes tells that there is problem with last HTTP(s) request/response
var ErrHTTPReqRes = fmt.Errorf("%w: something wrong with HTTP(s) request/response", ErrGdutils)

//ErrQJSON occurs when value could not be obtained from JSON
var ErrQJSON = fmt.Errorf("%w: could not obtain value from JSON", ErrJson)

//ErrPreservedData tells indices that there is some kind of error with scenario preserved data.
var ErrPreservedData = errors.New("preserved data error")
