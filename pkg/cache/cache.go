// Package cache holds definition of Cache used for storing and retrieving data.
package cache

import (
	"errors"
)

// ErrMissingKey occurs when cache doesn't have any value under given key.
var ErrMissingKey = errors.New("missing key")

// Cache is entity that has ability to store/retrieve arbitrary values.
type Cache interface {
	// Save preserve provided value under given key.
	Save(key string, value any)

	// GetSaved retrieve value from given key.
	GetSaved(key string) (any, error)

	// Reset turns cache into init state - clears all entries.
	Reset()

	// All returns all cache entries.
	All() map[string]any
}
