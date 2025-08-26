# Go concurrent non-blocking cache

[![Go Reference](https://pkg.go.dev/badge/github.com/ykalchevskiy/ccache.svg)](https://pkg.go.dev/github.com/ykalchevskiy/ccache)
[![Go Report Card](https://goreportcard.com/badge/github.com/ykalchevskiy/ccache)](https://goreportcard.com/report/github.com/ykalchevskiy/ccache)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ykalchevskiy/ccache)](https://golang.org/dl/)


## CCache

### Example

```go
package main

import "github.com/ykalchevskiy/ccache"

func main() {
    cache := ccache.New[string]()

    var v string

    v, _ = cache.Do("k1", func() (string, error) { return "v1", nil })
	println(v)                                                              // v1
    v, _ = cache.Do("k1", func() (string, error) { return "v111111", nil }) // v will not change
	println(v)                                                              // v1
}
```

### Example Concurrent

```go
package main

import (
    "sync"
    "sync/atomic"

    "github.com/ykalchevskiy/ccache"
)

func main() {
    cache := ccache.New[string]()

    wg := &sync.WaitGroup{}
	executionsCount := &atomic.Int32{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cache.Do("k2", func() (string, error) {
				executionsCount.Add(1)
				return "v2", nil
			})
			println(result) // v2
		}()
	}
	wg.Wait()

    println(executionsCount.Load()) // 1
}
```

## LRU

### Example Eviction

```go
package main

import "github.com/ykalchevskiy/ccache"

func main() {
    cache := ccache.NewLRU(2, ccache.OnEvictionFunc[string](func(key string) { println(key + " evicted") }))

    _, _ = cache.Do("k1", func() (string, error) { return "v1", nil }) //
    _, _ = cache.Do("k2", func() (string, error) { return "v2", nil }) //
    _, _ = cache.Do("k3", func() (string, error) { return "v3", nil }) // k1 evicted
    _, _ = cache.Do("k2", func() (string, error) { return "v2", nil }) //
    _, _ = cache.Do("k4", func() (string, error) { return "v4", nil }) // k3 evicted
}
```
