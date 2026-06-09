package cache

import (
	"context"
	"errors"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

var (
	localCache     *cache.Cache[any]
	singleSetCache singleflight.Group
)

func Init() {
	client := gocache.New(1*time.Minute, 2*time.Minute)
	storeImpl := gocache_store.NewGoCache(client)
	localCache = cache.New[any](storeImpl)
}

// GetByCache 先查本地缓存；miss 时通过 fn 加载并写回，singleflight 防止并发穿透。
func GetByCache[T any](key string, ttl time.Duration, fn func() (*T, error)) (*T, error) {
	ctx := context.Background()
	localKey := "cache:" + key

	if v, err := localCache.Get(ctx, localKey); err == nil {
		if data, ok := v.(*T); ok {
			return data, nil
		}
		return nil, errors.New("cache: type mismatch on hit")
	}

	v, err, _ := singleSetCache.Do(localKey, func() (any, error) {
		data, err := fn()
		if err != nil {
			return nil, err
		}
		_ = localCache.Set(ctx, localKey, data, store.WithExpiration(ttl))
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	data, ok := v.(*T)
	if !ok {
		return nil, errors.New("cache: type mismatch on load")
	}
	return data, nil
}
