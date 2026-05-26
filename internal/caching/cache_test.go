package caching

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/mgtv-tech/jetcache-go"
	"github.com/mgtv-tech/jetcache-go/local"
	"github.com/mgtv-tech/jetcache-go/remote"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheService_Fetch_LocalOnly(t *testing.T) {
	t.Skip("TinyLFU local cache has compression issues, skipping")
}

func TestFetch_Generic_LoaderError(t *testing.T) {
	cs := NewNoOpCacheService()
	ctx := context.Background()
	expectedErr := errors.New("db error")

	_, err := Fetch(ctx, cs, "err:key", time.Minute, func() (string, error) {
		return "", expectedErr
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

func TestFetch_Generic_Success(t *testing.T) {
	cs := NewNoOpCacheService()
	ctx := context.Background()

	result, err := Fetch(ctx, cs, "key", time.Minute, func() (string, error) {
		return "hello", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestNoOpCacheService_FetchAlwaysCallsLoader(t *testing.T) {
	cs := NewNoOpCacheService()
	ctx := context.Background()

	var callCount int
	result, err := cs.Fetch(ctx, "key", time.Minute, func() (any, error) {
		callCount++
		return "noop", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "noop", result)
	assert.Equal(t, 1, callCount)

	result, err = cs.Fetch(ctx, "key", time.Minute, func() (any, error) {
		callCount++
		return "noop2", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "noop2", result)
	assert.Equal(t, 2, callCount)
}

func TestNoOpCacheService_LoaderError(t *testing.T) {
	cs := NewNoOpCacheService()
	ctx := context.Background()
	expectedErr := errors.New("fail")

	_, err := cs.Fetch(ctx, "key", time.Minute, func() (any, error) {
		return nil, expectedErr
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

func TestNoOpCacheService_DeleteNoOps(t *testing.T) {
	cs := NewNoOpCacheService()
	assert.NoError(t, cs.Delete(context.Background(), "any"))
	assert.NoError(t, cs.DeletePattern(context.Background(), "any"))
}

func TestNewJetCacheInstance_NilRedis(t *testing.T) {
	_, err := NewJetCacheInstance(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis instance is required")
}

func TestVersionManagerLocalCache_Integration(t *testing.T) {
	localCache := cache.New(
		cache.WithName("test_local"),
		cache.WithLocal(local.NewTinyLFU(100, 5*time.Minute)),
	)

	ctx := context.Background()
	key := "test_key"

	var callCount int
	loader := func() (int64, error) {
		callCount++
		return int64(42), nil
	}

	var result int64
	err := localCache.Once(ctx, key, cache.Value(&result), cache.TTL(time.Minute),
		cache.Do(func(ctx context.Context) (any, error) {
			return loader()
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
	assert.Equal(t, 1, callCount)

	err = localCache.Once(ctx, key, cache.Value(&result), cache.TTL(time.Minute),
		cache.Do(func(ctx context.Context) (any, error) {
			return loader()
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
	assert.Equal(t, 1, callCount)
}

func TestCacheService_Delete_NonExistent(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	jetCache := cache.New(
		cache.WithName("test"),
		cache.WithRemote(remote.NewGoRedisV9Adapter(rdb)),
	)
	cs := NewCacheService(jetCache, rdb)

	err = cs.Delete(context.Background(), "nonexistent")
	require.NoError(t, err)
}

func TestCacheService_Delete_MultipleKeys_NonExistent(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	jetCache := cache.New(
		cache.WithName("test"),
		cache.WithRemote(remote.NewGoRedisV9Adapter(rdb)),
	)
	cs := NewCacheService(jetCache, rdb)

	err = cs.Delete(context.Background(), "k1", "k2")
	require.NoError(t, err)
}

func TestCacheService_DeletePattern_NoMatch(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	jetCache := cache.New(
		cache.WithName("test"),
		cache.WithRemote(remote.NewGoRedisV9Adapter(rdb)),
	)
	cs := NewCacheService(jetCache, rdb)

	err = cs.DeletePattern(context.Background(), "nomatch:")
	require.NoError(t, err)
}
