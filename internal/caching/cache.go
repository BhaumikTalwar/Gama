package caching

import (
	"fmt"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/mgtv-tech/jetcache-go"
	"github.com/mgtv-tech/jetcache-go/local"
	"github.com/mgtv-tech/jetcache-go/remote"
	"github.com/redis/go-redis/v9"
)

func NewJetCacheInstance(redisCl *redis.Client, cacheConfig *config.CacheConfig, notFoundErr ...error) (cache.Cache, error) {
	if redisCl == nil {
		return nil, fmt.Errorf("Redis Instance is required for the setup")
	}

	opts := []cache.Option{}
	opts = append(opts, cache.WithName(cacheConfig.CacheNamespace))
	opts = append(opts, cache.WithRemote(remote.NewGoRedisV9Adapter(redisCl)))

	if cacheConfig.EnableLocalCache {
		ttl := time.Duration(cacheConfig.LocalCacheTTL) * time.Minute
		localCache := local.NewFreeCache(local.Size(cacheConfig.LocalCacheSize), ttl)
		opts = append(opts, cache.WithLocal(localCache))
	}

	if len(notFoundErr) != 0 {
		opts = append(opts, cache.WithErrNotFound(notFoundErr[0]))
	}

	return cache.New(opts...), nil
}
