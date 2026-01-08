package workspace

import (
	"context"
	"sync"
	"time"
)

// DefaultCacheTTL is the default time-to-live for cached discovery results.
const DefaultCacheTTL = 5 * time.Second

// CachingDetector wraps a Detector and caches discovery results.
// This improves performance for large monorepos where multiple operations
// may need to detect modules within a short time window.
type CachingDetector struct {
	detector *Detector
	ttl      time.Duration

	mu        sync.RWMutex
	cache     *Context
	cacheRoot string
	cacheTime time.Time
}

// NewCachingDetector creates a new CachingDetector wrapping the given Detector.
// Uses DefaultCacheTTL if ttl is 0.
func NewCachingDetector(detector *Detector, ttl time.Duration) *CachingDetector {
	if ttl == 0 {
		ttl = DefaultCacheTTL
	}
	return &CachingDetector{
		detector: detector,
		ttl:      ttl,
	}
}

// DetectContext returns cached results if valid, otherwise performs detection.
//
// Cache is invalidated when:
//   - TTL has expired
//   - Root directory has changed
//   - InvalidateCache() was called
func (c *CachingDetector) DetectContext(ctx context.Context, root string) (*Context, error) {
	// Check cache first (read lock)
	c.mu.RLock()
	if c.isValidLocked(root) {
		result := c.cache
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Cache miss - perform detection (write lock)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.isValidLocked(root) {
		return c.cache, nil
	}

	// Perform actual detection
	result, err := c.detector.DetectContext(ctx, root)
	if err != nil {
		return nil, err
	}

	// Update cache
	c.cache = result
	c.cacheRoot = root
	c.cacheTime = time.Now()

	return result, nil
}

// DiscoverModules delegates to the underlying detector (not cached).
// Use DetectContext for cached discovery.
func (c *CachingDetector) DiscoverModules(ctx context.Context, root string) ([]*Module, error) {
	return c.detector.DiscoverModules(ctx, root)
}

// InvalidateCache clears the cached results.
// Call this after operations that modify .version files.
func (c *CachingDetector) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = nil
	c.cacheRoot = ""
	c.cacheTime = time.Time{}
}

// isValidLocked checks if cache is valid. Caller must hold at least a read lock.
func (c *CachingDetector) isValidLocked(root string) bool {
	if c.cache == nil {
		return false
	}
	if c.cacheRoot != root {
		return false
	}
	if time.Since(c.cacheTime) > c.ttl {
		return false
	}
	return true
}

// CacheInfo returns information about the current cache state.
// Useful for debugging and testing.
type CacheInfo struct {
	HasCache  bool
	Root      string
	Age       time.Duration
	ExpiresIn time.Duration
}

// GetCacheInfo returns the current cache state.
func (c *CachingDetector) GetCacheInfo() CacheInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return CacheInfo{HasCache: false}
	}

	age := time.Since(c.cacheTime)
	expiresIn := max(c.ttl-age, 0)

	return CacheInfo{
		HasCache:  true,
		Root:      c.cacheRoot,
		Age:       age,
		ExpiresIn: expiresIn,
	}
}
