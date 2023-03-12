package cache

import (
	"context"
	"github.com/gotomicro/ekit/list"
	"sync"
	"time"
)

type MaxMemoryCache struct {
	Cache
	max  int64 // 控制總量 （value）
	used int64

	// 雖然效能會影響，但是可以較好控制內存，不會有併發問題
	// 可以用原子操作替換，但是內存控制就不一定準確了
	mutex *sync.Mutex
	// 按照正常的設計，這邊需要的是一個接近 Java 的 LinkedHashMap 的結構
	keys *list.LinkedList[string]
}

//var _ Cache = (*MaxMemoryCache)(nil)

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	ret := &MaxMemoryCache{
		max:   max,
		Cache: cache,
		mutex: &sync.Mutex{},
		keys:  list.NewLinkedList[string](),
	}
	ret.Cache.OnEvicted(ret.evicted)
	return ret
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, val []byte,
	expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// 懶惰刪除檢查
	_, _ = m.Cache.LoadAndDelete(ctx, key)
	for m.used+int64(len(val)) > m.max {
		k, err := m.keys.Get(0)
		if err != nil {
			return err
		}
		_ = m.Cache.Delete(ctx, k)
	}
	err := m.Cache.Set(ctx, key, val, expiration)
	if err == nil {
		m.used = m.used + int64(len(val))
		_ = m.keys.Append(key)
	}
	return nil
}

func (m *MaxMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	// 加鎖，預防懶惰刪除的情況
	m.mutex.Lock()
	defer m.mutex.Unlock()

	val, err := m.Cache.Get(ctx, key)
	if err == nil {
		// 因為 LRU 策略
		// 所以在 linked list 先刪除，移到末尾
		m.deleteKey(key)
		_ = m.keys.Append(key)
	}
	return val, err
}

func (m *MaxMemoryCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Cache.Delete(ctx, key)
}

func (m *MaxMemoryCache) OnEvicted(f func(key string, val []byte)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Cache.OnEvicted(func(key string, val []byte) {
		m.evicted(key, val)
		f(key, val)
	})
}

func (m *MaxMemoryCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Cache.LoadAndDelete(ctx, key)
}

func (m *MaxMemoryCache) evicted(key string, val []byte) {
	m.used = m.used - int64(len(val))
	m.deleteKey(key)
}

func (m *MaxMemoryCache) deleteKey(key string) {
	for i := 0; i < m.keys.Len(); i++ {
		k, err := m.keys.Get(i)
		if err != nil {
			return
		}
		if k == key {
			_, _ = m.keys.Delete(i)
			return
		}
	}
}
