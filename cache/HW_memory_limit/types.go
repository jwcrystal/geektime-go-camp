package cache

import (
	"context"
	"time"
)

// Cache 遮蔽不同的緩存中間件的差異
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte,
		expiration time.Duration) error
	Delete(ctx context.Context, key string) error

	LoadAndDelete(ctx context.Context, key string) ([]byte, error)

	OnEvicted(func(key string, val []byte))
}

type CacheV1 interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte,
		expiration time.Duration) error
	Delete(ctx context.Context, key string) error

	LoadAndDelete(ctx context.Context, key string) ([]byte, error)

	Subscribe() <-chan Event
}

type Event struct {
	Key string
	Val any
	// 增删改查
	Type int
}
