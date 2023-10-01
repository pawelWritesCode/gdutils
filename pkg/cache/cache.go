package cache

import (
	"errors"
)

// ErrMissingKey occurs when cache doesn't have any value under given key.
var ErrMissingKey = errors.New("missing key")
