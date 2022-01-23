package cache

import (
	"fmt"
	"sync"
)

// ConcurrentCache is entity that has ability to store and retrieve arbitrary values.
// Safe for concurrent use.
type ConcurrentCache struct {
	buff sync.Map
}

// NewConcurrentCache returns pointer to ConcurrentCache safe for concurrent use
func NewConcurrentCache() *ConcurrentCache { return &ConcurrentCache{buff: sync.Map{}} }

func (c *ConcurrentCache) Save(key string, value interface{}) {
	c.buff.Store(key, value)
}

func (c *ConcurrentCache) GetSaved(key string) (interface{}, error) {
	val, ok := c.buff.Load(key)
	if ok == false {
		return val, fmt.Errorf("%w: %s", ErrMissingKey, key)
	}

	return val, nil
}

func (c *ConcurrentCache) Reset() {
	c.buff = sync.Map{}
}

func (c *ConcurrentCache) All() map[string]interface{} {
	tmpMap := make(map[string]interface{})
	c.buff.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}

		tmpMap[keyStr] = value

		return true
	})

	return tmpMap
}
