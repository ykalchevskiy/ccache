package ccache_test

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ykalchevskiy/ccache"
)

func ExampleCCache() {
	cache := ccache.New[string]()

	var v string

	v, _ = cache.Do("k1", func() (string, error) { return "v1", nil })
	fmt.Println(v)                                                          // v1
	v, _ = cache.Do("k1", func() (string, error) { return "v111111", nil }) // v will not change
	fmt.Println(v)                                                          // v1

	// Output:
	// v1
	// v1
}

func ExampleCCache_concurrent() {
	cache := ccache.New[string]()

	wg := &sync.WaitGroup{}
	executionsCount := &atomic.Int32{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cache.Do("key", func() (string, error) {
				executionsCount.Add(1)
				return "value", nil
			})
			fmt.Println(result, err) // value <nil>
		}()
	}
	wg.Wait()

	fmt.Println(executionsCount.Load()) // 1

	// Output:
	// value <nil>
	// value <nil>
	// value <nil>
	// value <nil>
	// value <nil>
	// 1
}

func ExampleLRU() {
	cache := ccache.MustLRU(2, ccache.OnEvictionFunc[string](func(key string) { fmt.Println(key + " evicted") }))

	_, _ = cache.Do("k1", func() (string, error) { return "v1", nil }) //
	_, _ = cache.Do("k2", func() (string, error) { return "v2", nil }) //
	_, _ = cache.Do("k3", func() (string, error) { return "v3", nil }) // k1 evicted
	_, _ = cache.Do("k2", func() (string, error) { return "v2", nil }) //
	_, _ = cache.Do("k4", func() (string, error) { return "v4", nil }) // k3 evicted

	// Output:
	// k1 evicted
	// k3 evicted
}
