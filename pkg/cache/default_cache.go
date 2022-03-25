package cache

import "fmt"

// DefaultCache is entity that has ability to store and retrieve arbitrary values.
// Not safe for concurrent use.
type DefaultCache struct {
	buff map[string]any
}

// NewDefaultCache returns pointer to DefaultCache not safe for concurrent use
func NewDefaultCache() *DefaultCache { return &DefaultCache{buff: map[string]any{}} }

// Save preserve value under given key in DefaultCache.
func (c *DefaultCache) Save(key string, value any) {
	c.buff[key] = value
}

// GetSaved returns preserved value if present, error otherwise.
func (c *DefaultCache) GetSaved(key string) (any, error) {
	val, ok := c.buff[key]

	if ok == false {
		return val, fmt.Errorf("%w: %s", ErrMissingKey, key)
	}

	return val, nil
}

// Reset turns cache into initial state - - clears all entries.
func (c *DefaultCache) Reset() {
	c.buff = map[string]any{}
}

// All returns all current cache data.
func (c *DefaultCache) All() map[string]any {
	return c.buff
}
