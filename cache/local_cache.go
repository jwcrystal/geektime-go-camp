package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: key not found")
	//errKeyExpired = errors.New("cache：key was expired")
)

type item struct {
	val      any
	deadline time.Time
}

type BuildInMapCache struct {
	data map[string]*item
	//data sync.Map // 無法做到較為精細的控制
	mu        sync.RWMutex
	onEvicted func(key string, val any)
	close     chan struct{}
}

func (i item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}

type BuildInMapCacheOption func(cache *BuildInMapCache)

func NewBuildInMapCache(interval time.Duration, opts ...BuildInMapCacheOption) *BuildInMapCache {
	res := &BuildInMapCache{
		data:      make(map[string]*item, 100),
		onEvicted: func(key string, val any) {},
		close:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(res)
	}

	// 檢查過期, goroutine 輪詢
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				res.mu.Lock()
				i := 0
				for key, val := range res.data {
					// 遍歷緩存筆數控制
					if i > 10000 {
						return
					}
					// 已設置過期時間，且已過期了
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					i++
				}
				res.mu.Unlock()
			case <-res.close:
				return
			}
		}
	}()

	return res
}

func BuildInMapCacheWithOnEvictedCallback(f func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = f
	}
}

func (b *BuildInMapCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.set(ctx, key, value, expiration)
}

func (b *BuildInMapCache) set(ctx context.Context, key string, value any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		//time.AfterFunc(expiration, func() {
		//	delete(b.data, key)
		//})
		dl = time.Now().Add(expiration)
	}
	b.data[key] = &item{val: value, deadline: dl}
	return nil
}

func (b *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	// Get 找不到 key 情況
	// 1. 原本就沒有 key
	// 2. key 在拿到時候就過期了
	// 3. 要拿 key 時候，被其他人更新過期時間導致過期了
	b.mu.RLock()
	itm, ok := b.data[key]
	b.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key) // 沒有這個 key，所以找不到
	}
	now := time.Now()
	if itm.deadlineBefore(now) {
		// 確認狀況三
		b.mu.Lock()
		defer b.mu.Unlock()
		itm, ok = b.data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
		// 如果沒有人刪除，就再檢查一次 key 是否過期
		// 過期就刪除，等輪詢太慢了, 會浪費資源
		if itm.deadlineBefore(now) {
			b.delete(key)
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
	}
	return itm.val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delete(key)
	return nil
}

func (b *BuildInMapCache) delete(key string) {
	itm, ok := b.data[key]
	if !ok {
		return
	}
	delete(b.data, key)
	b.onEvicted(key, itm.val)
}

func (b *BuildInMapCache) Close() error {
	select {
	case b.close <- struct{}{}:
		return nil
	default:
		return errors.New("重複關閉")
	}
}
