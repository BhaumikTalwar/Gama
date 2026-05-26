package caching

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/mgtv-tech/jetcache-go"
	"github.com/mgtv-tech/jetcache-go/local"
	"github.com/mgtv-tech/jetcache-go/remote"
	"github.com/redis/go-redis/v9"
)

func NewJetCacheInstance(redisCl *redis.Client, cacheConfig *config.CacheConfig, notFoundErr ...error) (cache.Cache, error) {
	if redisCl == nil {
		return nil, fmt.Errorf("redis instance is required for the setup")
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

type CacheService interface {
	Fetch(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error)
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
}

type cacheService struct {
	client cache.Cache
	redis  redis.Cmdable
}

func NewCacheService(client cache.Cache, redisClient *redis.Client) CacheService {
	return &cacheService{
		client: client,
		redis:  redisClient,
	}
}

func (c *cacheService) Fetch(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	var result any
	err := c.client.Once(ctx, key, cache.Value(&result), cache.TTL(ttl),
		cache.Do(func(ctx context.Context) (any, error) {
			return loader()
		}))
	if err != nil {
		return nil, fmt.Errorf("cache fetch failed: %w", err)
	}
	return result, nil
}

func (c *cacheService) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if err := c.client.Delete(ctx, key); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", key, err)
		}
	}
	return nil
}

func (c *cacheService) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	matchPattern := pattern
	if !strings.HasSuffix(pattern, "*") {
		matchPattern = pattern + "*"
	}

	for {
		keys, nextCursor, err := c.redis.Scan(ctx, cursor, matchPattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			for _, key := range keys {
				if err := c.client.Delete(ctx, key); err != nil {
					return fmt.Errorf("failed to delete key %s: %w", key, err)
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

type noOpCacheService struct{}

func NewNoOpCacheService() CacheService {
	return &noOpCacheService{}
}

func (n *noOpCacheService) Fetch(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	return loader()
}

func (n *noOpCacheService) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n *noOpCacheService) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

type FetchFunc func(ctx context.Context) (any, error)

func Fetch[T any](ctx context.Context, cs CacheService, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	result, err := cs.Fetch(ctx, key, ttl, func() (any, error) {
		return loader()
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result.(T), nil
}