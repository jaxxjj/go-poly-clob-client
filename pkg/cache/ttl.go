// Package cache provides a concurrent-safe TTL cache.
package cache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value   V
	expires time.Time
}

// TTL is a concurrent-safe cache with per-key expiration.
type TTL[V any] struct {
	mu  sync.RWMutex
	ttl time.Duration
	m   map[string]entry[V]
}

// New creates a TTL cache with the given default duration.
func New[V any](ttl time.Duration) *TTL[V] {
	return &TTL[V]{
		ttl: ttl,
		m:   make(map[string]entry[V]),
	}
}

// Get retrieves a value if it exists and has not expired.
func (c *TTL[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expires) {
		var zero V
		return zero, false
	}
	return e.value, true
}

// Set stores a value with the default TTL.
func (c *TTL[V]) Set(key string, value V) {
	c.mu.Lock()
	c.m[key] = entry[V]{value: value, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Clear removes specific keys, or all keys if none are specified.
func (c *TTL[V]) Clear(keys ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(keys) == 0 {
		c.m = make(map[string]entry[V])
		return
	}
	for _, k := range keys {
		delete(c.m, k)
	}
}
