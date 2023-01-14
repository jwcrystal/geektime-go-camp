package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuildInMapCache_Get(t *testing.T) {
	testCases := []struct {
		name    string
		cache   func() *BuildInMapCache
		key     string
		wantErr error
		wantVal any
	}{
		{
			name: "key not found",
			key:  "not existed key",
			cache: func() *BuildInMapCache {
				return NewBuildInMapCache(time.Second * 10)
			},
			wantErr: fmt.Errorf("%w, key: %s", errKeyNotFound, "not existed key"),
		},
		{
			name: "key expired",
			key:  "expired key",
			cache: func() *BuildInMapCache {
				c := NewBuildInMapCache(time.Second * 10)
				err := c.Set(context.Background(), "expired key", "123", time.Second)
				require.NoError(t, err)
				time.Sleep(time.Second * 2)
				return c
			},
			wantErr: fmt.Errorf("%w, key: %s", errKeyNotFound, "expired key"),
		},
		{
			name: "get key",
			key:  "key1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(time.Second)
				err := res.Set(context.Background(), "key1", 123, time.Second)
				require.NoError(t, err)
				return res
			},
			wantVal: 123,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			itm, err := tc.cache().Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, itm)
		})
	}
}

func TestBuildInMapCache_Loop(t *testing.T) {
	count := 0
	cache := NewBuildInMapCache(time.Second,
		BuildInMapCacheWithOnEvictedCallback(func(key string, val any) {
			count++
		}))
	err := cache.Set(context.Background(), "key", "123", time.Second)
	assert.NoError(t, err)
	time.Sleep(time.Second * 3)
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	// 不能 Get，會把 key 刪除
	//itm, err := cache.Get(context.Background(), "key")
	_, ok := cache.data["key"]
	assert.False(t, ok)
	assert.Equal(t, 1, count)
}
