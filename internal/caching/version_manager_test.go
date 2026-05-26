package caching

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/mgtv-tech/jetcache-go"
	"github.com/mgtv-tech/jetcache-go/local"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupVersionManager(t *testing.T) (*VersionManager, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	vm := NewVersionManager(rdb)
	return vm, mr
}

func TestVersionManager_GetVersion_InitializesToOne(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()
	version, err := vm.GetVersion(ctx, "user", "123")
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)
}

func TestVersionManager_GetVersion_ReturnsStoredVersion(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()

	_ = mr.Set("version:user:123", "5")

	version, err := vm.GetVersion(ctx, "user", "123")
	require.NoError(t, err)
	assert.Equal(t, int64(5), version)
}

func TestVersionManager_Touch_IncrementsVersion(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()

	v1, _ := vm.GetVersion(ctx, "user", "123")
	assert.Equal(t, int64(1), v1)

	err := vm.Touch(ctx, "user", "123")
	require.NoError(t, err)

	v2, _ := vm.GetVersion(ctx, "user", "123")
	assert.Equal(t, int64(2), v2)
}

func TestVersionManager_TouchEntity_IncrementsGlobalVersion(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()

	v1, _ := vm.GetEntityVersions(ctx, "user")
	assert.Equal(t, int64(1), v1)

	err := vm.TouchEntity(ctx, "user")
	require.NoError(t, err)

	v2, _ := vm.GetEntityVersions(ctx, "user")
	assert.Equal(t, int64(2), v2)
}

func TestVersionManager_TouchMultiple_IncrementsMultiple(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()

	_, _ = vm.GetVersion(ctx, "user", "1")
	_, _ = vm.GetVersion(ctx, "role", "admin")
	_, _ = vm.GetVersion(ctx, "role", "global")

	err := vm.TouchMultiple(ctx, map[string][]string{
		"user": {"1"},
		"role": {"admin", "global"},
	})
	require.NoError(t, err)

	vu, _ := vm.GetVersion(ctx, "user", "1")
	vr, _ := vm.GetVersion(ctx, "role", "admin")
	vg, _ := vm.GetVersion(ctx, "role", "global")

	assert.Equal(t, int64(2), vu)
	assert.Equal(t, int64(2), vr)
	assert.Equal(t, int64(2), vg)
}

func TestVersionManager_TouchEntityMultiple_IncrementsMultipleEntities(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()

	_ = vm.TouchEntity(ctx, "user")
	_ = vm.TouchEntity(ctx, "role")

	err := vm.TouchEntityMultiple(ctx, []string{"user", "role"})
	require.NoError(t, err)

	vu, _ := vm.GetEntityVersions(ctx, "user")
	vr, _ := vm.GetEntityVersions(ctx, "role")

	assert.Equal(t, int64(2), vu)
	assert.Equal(t, int64(2), vr)
}

func TestVersionManager_GetMultiple_ReturnsAllVersions(t *testing.T) {
	vm, mr := setupVersionManager(t)
	defer mr.Close()

	ctx := context.Background()

	_ = mr.Set("version:user:1", "3")
	_ = mr.Set("version:role:admin", "5")

	versions, err := vm.GetMultiple(ctx, map[string][]string{
		"user": {"1", "2"},
		"role": {"admin"},
	})
	require.NoError(t, err)

	assert.Equal(t, int64(3), versions["user"]["1"])
	assert.Equal(t, int64(1), versions["user"]["2"]) // not set, defaults to 1
	assert.Equal(t, int64(5), versions["role"]["admin"])
}

func TestVersionManager_localCacheOnly(t *testing.T) {
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
