package judging

import (
	"context"
	"encoding/json"
	"sync"
)

// CacheKey identifies one cached generation by protocol, instance, step, and
// prompt digest. Prompt digest is included so a prompt change is a new
// population even when protocol and instance are unchanged.
type CacheKey struct {
	ProtocolDigest string
	InstanceDigest string
	Step           string
	PromptDigest   string
}

// Cache stores and retrieves generated outputs by CacheKey. Caching improves
// reproducibility and cost control; it does not count as reliability
// measurement, so audit runners use a cache-bypass path.
type Cache interface {
	Load(ctx context.Context, key CacheKey, out any) (bool, error)
	Store(ctx context.Context, key CacheKey, value any) error
}

// NoopCache is a Cache that never stores or loads.
type NoopCache struct{}

// Load always reports a miss.
func (NoopCache) Load(_ context.Context, _ CacheKey, _ any) (bool, error) { return false, nil }

// Store discards the value.
func (NoopCache) Store(_ context.Context, _ CacheKey, _ any) error { return nil }

// MemoryCache is an in-process Cache for tests and ephemeral runs. It is safe
// for concurrent use. It is not durable.
type MemoryCache struct {
	mu    sync.Mutex
	store map[CacheKey]any
}

// NewMemoryCache returns an empty in-memory cache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{store: make(map[CacheKey]any)}
}

// Load retrieves a value by key and JSON-decodes it into out (a pointer).
// It returns (false, nil) on a miss.
func (c *MemoryCache) Load(_ context.Context, key CacheKey, out any) (bool, error) {
	c.mu.Lock()
	v, ok := c.store[key]
	c.mu.Unlock()
	if !ok {
		return false, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return false, err
	}
	return true, nil
}

// Store saves a value by key.
func (c *MemoryCache) Store(_ context.Context, key CacheKey, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return nil
}
