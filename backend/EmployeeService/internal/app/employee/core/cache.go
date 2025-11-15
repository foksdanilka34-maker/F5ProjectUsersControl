package employee

import (
	"sync"
	"time"
)

type cacheEntry[T any] struct {
	value   *T
	expires time.Time
}

type ttlCache[T any] struct {
	mu   sync.RWMutex
	data map[string]cacheEntry[T]
	ttl  time.Duration
}

func newTTLCache[T any](ttl time.Duration) *ttlCache[T] {
	return &ttlCache[T]{
		data: make(map[string]cacheEntry[T]),
		ttl:  ttl,
	}
}

func (c *ttlCache[T]) Get(key string) (*T, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expires) {
		c.Delete(key)
		return nil, false
	}
	return entry.value, true
}

func (c *ttlCache[T]) Set(key string, value *T) {
	if c == nil || value == nil {
		return
	}

	c.mu.Lock()
	c.data[key] = cacheEntry[T]{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *ttlCache[T]) Delete(key string) {
	if c == nil {
		return
	}

	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}
