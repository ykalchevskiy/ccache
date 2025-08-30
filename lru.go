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

type OptionLRU[T any] func(*LRU[T])

func OnEvictionFunc[T any](f func(key string)) OptionLRU[T] {
	return func(c *LRU[T]) {
		c.onEviction = f
	}
}

type LRU[T any] struct {
	size       int
	onEviction func(key string)

	mu sync.Mutex
	m  map[string]*entryLRU[T]
	l  *list.List
}

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

func MustLRU[T any](size int, opts ...OptionLRU[T]) *LRU[T] {
    c, err := NewLRU(size, opts...)
    if err != nil {
        panic(err)
    }
    return c
}

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
