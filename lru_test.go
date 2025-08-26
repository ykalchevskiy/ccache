package ccache_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ykalchevskiy/ccache"
)

func TestLRU(t *testing.T) {
	var evictedKeys []string
	cache := ccache.NewLRU(2, ccache.OnEvictionFunc[int](func(key string) {
		evictedKeys = append(evictedKeys, key)
	}))

	var result int
	var err error

	result, err = cache.Do("k1", func() (int, error) { return 1, nil })
	require(t, nil, err)
	require(t, 1, result)
	requireSlice(t, nil, evictedKeys)

	result, err = cache.Do("k1", func() (int, error) { return 111111, nil }) // another value will not be used
	require(t, nil, err)
	require(t, 1, result)
	requireSlice(t, nil, evictedKeys)

	result, err = cache.Do("k2", func() (int, error) { return 2, nil })
	require(t, nil, err)
	require(t, 2, result)
	requireSlice(t, nil, evictedKeys)

	result, err = cache.Do("k3", func() (int, error) { return 3, nil })
	require(t, nil, err)
	require(t, 3, result)
	requireSlice(t, []string{"k1"}, evictedKeys)
	evictedKeys = nil

	result, err = cache.Do("k2", func() (int, error) { return 2, nil })
	require(t, nil, err)
	require(t, 2, result)
	requireSlice(t, nil, evictedKeys)

	result, err = cache.Do("k4", func() (int, error) { return 4, nil })
	require(t, nil, err)
	require(t, 4, result)
	requireSlice(t, []string{"k3"}, evictedKeys)
	evictedKeys = nil
}

func TestLRU_concurrent(t *testing.T) {
	evictedKeysCount := &atomic.Int32{}
	cache := ccache.NewLRU(1, ccache.OnEvictionFunc[int](func(key string) {
		evictedKeysCount.Add(1)
	}))

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

	require(t, int32(0), evictedKeysCount.Load())
	require(t, int32(1), storeCount.Load())
}
