// Package cache holds definition of Cache used for storing and retrieving data.
package cache

import (
	"errors"
	"fmt"
)

// ErrMissingKey occurs when cache doesn't have any value under given key.
var ErrMissingKey = errors.New("missing key")

// Cache is entity that has ability to store/retrieve arbitrary values.
type Cache interface {
	// Save preserve provided value under given key.
	Save(key string, value interface{})

	// GetSaved retrieve value from given key.
	GetSaved(key string) (interface{}, error)

	// Reset turns cache into init state - clears all entries.
	Reset()

	// All returns all cache entries.
	All() map[string]interface{}
}

// DefaultCache is entity that has ability to store and retrieve arbitrary values.
type DefaultCache struct {
	buff map[string]interface{}
}

func New() *DefaultCache {
	return &DefaultCache{buff: map[string]interface{}{}}
}

// Save preserve value under given key in DefaultCache.
func (c *DefaultCache) Save(key string, value interface{}) {
	c.buff[key] = value
}

// GetSaved returns preserved value if present, error otherwise.
func (c *DefaultCache) GetSaved(key string) (interface{}, error) {
	val, ok := c.buff[key]

	if ok == false {
		return val, fmt.Errorf("%w: %s", ErrMissingKey, key)
	}

	return val, nil
}

// Reset turns cache into initial state - - clears all entries.
func (c *DefaultCache) Reset() {
	c.buff = map[string]interface{}{}
}

// All returns all current cache data.
func (c *DefaultCache) All() map[string]interface{} {
	return c.buff
}
