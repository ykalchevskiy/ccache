package ccache_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ykalchevskiy/ccache"
)

func TestLRU(t *testing.T) {
	var evictedKeys []string
	cache := ccache.MustLRU(2, ccache.OnEvictionFunc[int](func(key string) {
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
	cache := ccache.MustLRU(1, ccache.OnEvictionFunc[int](func(key string) {
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

func TestLRU_sizeValidation(t *testing.T) {
	tests := []struct {
		name          string
		size          int
		expectedError string
	}{
		{
			name:          "zero size",
			size:          0,
			expectedError: "ccache: size must be greater than 0, got 0",
		},
		{
			name:          "negative size",
			size:          -1,
			expectedError: "ccache: size must be greater than 0, got -1",
		},
		{
			name:          "valid size",
			size:          1,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_NewLRU", func(t *testing.T) {
			_, err := ccache.NewLRU[string](tt.size)
			if err == nil && tt.expectedError != "" {
				t.Errorf("NewLRU() expected error %q, got nil", tt.expectedError)
			}
			if err != nil {
				if tt.expectedError == "" {
					t.Errorf("NewLRU() unexpected error: %v", err)
				} else if err.Error() != tt.expectedError {
					t.Errorf("NewLRU() error message = %q, want %q", err.Error(), tt.expectedError)
				}
			}
		})

		t.Run(tt.name+"_MustLRU", func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && tt.expectedError != "" {
					t.Errorf("MustLRU() expected panic with %q, got nil", tt.expectedError)
				}
				if r != nil {
					if tt.expectedError == "" {
						t.Errorf("MustLRU() unexpected panic: %v", r)
					} else if err, ok := r.(error); !ok {
						t.Errorf("MustLRU() panic value is not an error: %v", r)
					} else if err.Error() != tt.expectedError {
						t.Errorf("MustLRU() panic message = %q, want %q", err.Error(), tt.expectedError)
					}
				}
			}()
			_ = ccache.MustLRU[string](tt.size)
		})
	}
}
