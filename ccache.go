package ccache

import (
	"sync"
)

type entry[T any] struct {
	result T
	err    error
	ready  chan struct{}
}

// CCache implements a concurrent cache that memoizes function results.
// It is safe for concurrent use and has no size limit.
// Multiple goroutines can request the same key concurrently,
// but the function will only be executed once.
type CCache[T any] struct {
	mu sync.Mutex
	m  map[string]*entry[T]
}

// New creates a new concurrent cache.
// The cache has no size limit and will grow as needed.
func New[T any]() *CCache[T] {
	c := &CCache[T]{
		m: make(map[string]*entry[T]),
	}

	return c
}

// Do executes and memoizes the result of function f with the given key.
// If the key exists in the cache, the cached result is returned.
// If multiple goroutines call Do with the same key concurrently,
// only one execution of f will occur, and all callers will receive the same result.
func (c *CCache[T]) Do(key string, f func() (T, error)) (T, error) {
	c.mu.Lock()
	e, ok := c.m[key]
	if !ok {
		e = &entry[T]{
			ready: make(chan struct{}),
		}
		c.m[key] = e
		c.mu.Unlock()

		e.result, e.err = f()
		close(e.ready)
	} else {
		c.mu.Unlock()
		<-e.ready
	}

	return e.result, e.err
}
