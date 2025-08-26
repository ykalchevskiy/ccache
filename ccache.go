package ccache

import (
	"sync"
)

type entry[T any] struct {
	result T
	err    error
	ready  chan struct{}
}

type CCache[T any] struct {
	mu sync.Mutex
	m  map[string]*entry[T]
}

func New[T any]() *CCache[T] {
	c := &CCache[T]{
		m: make(map[string]*entry[T]),
	}

	return c
}

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
