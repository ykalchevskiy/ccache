package ccache

import (
	"sync"
)

// CCache implements a concurrent cache that memoizes function results.
// It is safe for concurrent use and has no size limit.
// Multiple goroutines can request the same key concurrently,
// but the function will only be executed once.
type CCache[T any] struct {
	m sync.Map
}

// New creates a new concurrent cache.
// The cache has no size limit and will grow as needed.
func New[T any]() *CCache[T] {
	return &CCache[T]{}
}

// Do executes and memoizes the result of function f with the given key.
// If the key exists in the cache, the cached result is returned.
// If multiple goroutines call Do with the same key concurrently,
// only one execution of f will occur, and all callers will receive the same result.
func (c *CCache[T]) Do(key string, f func() (T, error)) (T, error) {
	v, _ := c.m.LoadOrStore(key, sync.OnceValues(f))

	return v.(func() (T, error))()
}
