package ccache_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ykalchevskiy/ccache"
)

func TestCCache(t *testing.T) {
	cache := ccache.New[int]()

	var result int
	var err error

	result, err = cache.Do("k1", func() (int, error) { return 1, nil })
	require(t, nil, err)
	require(t, 1, result)

	result, err = cache.Do("k1", func() (int, error) { return 111111, nil }) // another value will not be used
	require(t, nil, err)
	require(t, 1, result)

	result, err = cache.Do("k2", func() (int, error) { return 2, nil })
	require(t, nil, err)
	require(t, 2, result)

	result, err = cache.Do("k3", func() (int, error) { return 3, nil })
	require(t, nil, err)
	require(t, 3, result)
}

func TestCCache_concurrent(t *testing.T) {
	cache := ccache.New[int]()

	wg := &sync.WaitGroup{}
	storeCount := &atomic.Int32{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cache.Do("k1", func() (int, error) {
				storeCount.Add(1)
				return 1, nil
			})
			assert(t, nil, err)
			assert(t, 1, result)
		}()
	}
	wg.Wait()

	require(t, int32(1), storeCount.Load())
}
