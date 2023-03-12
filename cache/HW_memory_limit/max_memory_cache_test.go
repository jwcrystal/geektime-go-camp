package cache

import (
	"context"
	"errors"
	"github.com/gotomicro/ekit/list"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var errNotFound error = errors.New("not found")

func TestMaxMemoryCache_Set(t *testing.T) {
	testCases := []struct {
		name  string
		cache func() *MaxMemoryCache

		key string
		val []byte

		wantKeys []string
		wantErr  error
		wantUsed int64
	}{
		{
			// 不會觸發淘汰
			name: "not exist",
			cache: func() *MaxMemoryCache {
				ret := NewMaxMemoryCache(10, &mockCache{data: map[string][]byte{}})
				return ret
			},
			key:      "key1",
			val:      []byte("value1"),
			wantKeys: []string{"key1"},
			wantUsed: 6,
		},
		{
			name: "override-used-incr",
			cache: func() *MaxMemoryCache {
				ret := NewMaxMemoryCache(10, &mockCache{data: map[string][]byte{
					"key1": []byte("value1"),
				}})
				ret.keys = list.NewLinkedListOf[string]([]string{"key1"})
				ret.used = 6
				return ret
			},
			key:      "key1",
			val:      []byte("value1-new"),
			wantKeys: []string{"key1"},
			wantUsed: 10,
		},
		{
			name: "override-used-decr",
			cache: func() *MaxMemoryCache {
				ret := NewMaxMemoryCache(10, &mockCache{data: map[string][]byte{
					"key1": []byte("value1"),
				}})
				ret.keys = list.NewLinkedListOf[string]([]string{"key1"})
				ret.used = 6
				return ret
			},
			key:      "key1",
			val:      []byte("val1"),
			wantKeys: []string{"key1"},
			wantUsed: 4,
		},
		{
			//進行淘汰
			name: "delete",
			cache: func() *MaxMemoryCache {
				ret := NewMaxMemoryCache(10, &mockCache{data: map[string][]byte{
					"key1": []byte("value1"),
					"key2": []byte("value2"),
					"key3": []byte("value3"),
				}})
				ret.keys = list.NewLinkedListOf[string]([]string{"key3"})
				ret.used = 6
				return ret
			},
			key:      "key4",
			val:      []byte("value4"),
			wantKeys: []string{"key4"},
			wantUsed: 6,
		},
		{
			//進行淘汰多次
			name: "multi-delete",
			cache: func() *MaxMemoryCache {
				ret := NewMaxMemoryCache(25, &mockCache{data: map[string][]byte{
					"key1": []byte("value1"),
					"key2": []byte("value2"),
					"key3": []byte("value3"),
				}})
				ret.keys = list.NewLinkedListOf[string]([]string{"key1", "key2", "key3"})
				ret.used = 18
				return ret
			},
			key:      "key4",
			val:      []byte("value4, value4, value4"),
			wantKeys: []string{"key4"},
			wantUsed: 22,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			err := cache.Set(context.Background(), tc.key, tc.val, time.Minute)
			assert.Equal(t, tc.wantKeys, cache.keys.AsSlice())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUsed, cache.used)
		})
	}
}

func TestMaxMemoryCache_Get(t *testing.T) {
	testCases := []struct {
		name  string
		cache func() *MaxMemoryCache

		key string

		wantKeys []string
		wantErr  error
	}{
		{
			name: "not exist",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{})
				return res
			},
			key:      "key1",
			wantKeys: []string{},
			wantErr:  errNotFound,
		},
		{
			name: "exist",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{data: map[string][]byte{
					"key1": []byte("value1"),
				}})
				return res
			},
			key:      "key1",
			wantKeys: []string{"key1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			_, err := cache.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantKeys, cache.keys.AsSlice())
		})
	}
}

type mockCache struct {
	f    func(key string, val []byte)
	data map[string][]byte
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if ok {
		return val, nil
	}
	return nil, errNotFound
}

func (m *mockCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	m.data[key] = val
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	val, ok := m.data[key]
	//delete(m.data, key)
	if ok {
		m.f(key, val)
	}
	return nil
}

func (m *mockCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if ok {
		m.f(key, val)
		return val, nil
	}
	return nil, errNotFound
}

func (m *mockCache) OnEvicted(f func(key string, val []byte)) {
	m.f = f
}
