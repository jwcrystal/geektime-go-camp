# 緩存內存總量控制

## 需求

- 功能性需求
  - 為緩存框架提供控制內存使用量的功能，用戶能夠控制內存使用量不超過一個閾值
    - 如 1G 等。這裡**只需要計算鍵值對中值的大小**，忽略鍵的大小
  - 可以提供不同淘汰算法的實現，至少要提供 LRU 淘汰策略
- 非功能需求
  - 用戶可以選擇不同的緩存底層實現
  - 這個功能必須是無侵入式地，即不影響已有緩存實現

## 思考方向

- 使用裝飾器模式，封裝原先的 Cache 實現，而外增加控制條件

```go
type MaxMemoryCache struct {
	Cache
	max int64
	used int64
}

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		max: max,
		Cache: cache,
	}
	res.Cache.OnEvicted(func(key string, val []byte) {
		// 透過回調機制，進行控制
	})
	return res
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, val []byte,
	expiration time.Duration) error {
	// Set 時候判斷 key有沒有存在及是否過期（檢查懶惰刪除）
  // 再判斷是否有超出內存總量閾值
	// 超出閾值，根據 LRU 策略，刪除最少用的 key
	return m.Cache.Set(ctx, key, val, expiration)
}

```

