package gdutils

import "errors"

//ErrJson tells that value has invalid JSON format.
var ErrJson = errors.New("invalid JSON format")

//ErrResponseCode tells that response had invalid response code.
var ErrResponseCode = errors.New("invalid response code")

//ErrJsonNode tells that there is some kind of error with json node.
var ErrJsonNode = errors.New("invalid JSON node")

//ErrPreservedData tells indices that there is some kind of error with scenario preserved data.
var ErrPreservedData = errors.New("preserved data error")
