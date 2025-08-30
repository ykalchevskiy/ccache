package ccache

import (
	"container/list"
	"fmt"
	"sync"
)

type entryLRU[T any] struct {
	result T
	err    error
	ready  chan struct{}

	element *list.Element
}

// OptionLRU is a function type that configures an LRU cache.
// Options are applied in the order they are provided.
type OptionLRU[T any] func(*LRU[T])

// OnEvictionFunc returns an option that sets a callback function
// which is called when an entry is evicted from the cache.
// The callback receives the key of the evicted entry.
func OnEvictionFunc[T any](f func(key string)) OptionLRU[T] {
	return func(c *LRU[T]) {
		c.onEviction = f
	}
}

// LRU implements a concurrent LRU (Least Recently Used) cache that memoizes function results.
// It maintains a maximum size and evicts the least recently used entries when the size is exceeded.
// It is safe for concurrent use.
type LRU[T any] struct {
	size       int
	onEviction func(key string)

	mu sync.Mutex
	m  map[string]*entryLRU[T]
	l  *list.List
}

// NewLRU creates a new LRU cache with the specified maximum size.
// The size must be greater than 0.
// Optional configuration can be provided via opts.
// Returns an error if size is less than or equal to 0.
func NewLRU[T any](size int, opts ...OptionLRU[T]) (*LRU[T], error) {
	if size <= 0 {
		return nil, fmt.Errorf("ccache: size must be greater than 0, got %d", size)
	}

	c := &LRU[T]{
		size: size,

		m: make(map[string]*entryLRU[T]),
		l: list.New(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// MustLRU is like NewLRU but panics if size is invalid.
// It simplifies initialization of global variables holding LRU caches.
func MustLRU[T any](size int, opts ...OptionLRU[T]) *LRU[T] {
	c, err := NewLRU(size, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// Do executes and memoizes the result of function f with the given key.
// If the key exists in the cache, the cached result is returned and the entry
// is moved to the front of the LRU list.
// If the cache is at capacity when adding a new entry, the least recently used
// entry is evicted.
// If multiple goroutines call Do with the same key concurrently,
// only one execution of f will occur, and all callers will receive the same result.
func (c *LRU[T]) Do(key string, f func() (T, error)) (T, error) {
	c.mu.Lock()
	e, ok := c.m[key]
	if !ok {
		if len(c.m) == c.size {
			elemToEvict := c.l.Back()
			c.l.Remove(elemToEvict)
			keyToEvict := elemToEvict.Value.(string)
			delete(c.m, keyToEvict)
			if c.onEviction != nil {
				c.onEviction(keyToEvict)
			}
		}
		elem := c.l.PushFront(key)
		e = &entryLRU[T]{
			ready:   make(chan struct{}),
			element: elem,
		}
		c.m[key] = e
		c.mu.Unlock()

		e.result, e.err = f()
		close(e.ready)
	} else {
		c.l.MoveToFront(e.element)
		c.mu.Unlock()
		<-e.ready
	}

	return e.result, e.err
}
