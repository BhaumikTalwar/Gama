package caching

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mgtv-tech/jetcache-go"
	"github.com/mgtv-tech/jetcache-go/local"
	"github.com/redis/go-redis/v9"
)

var errVersionNotFound = errors.New("version record not found")

const (
	versionPrefix     = "version"
	localCacheSize    = 10000
	localCacheTTL     = 10 * time.Minute
	versionKeyTTL     = 30 * 24 * time.Hour
	entityGlobalSuffix = "global"
)

type VersionManager struct {
	rdb        redis.Cmdable
	localCache cache.Cache
}

func NewVersionManager(rdb redis.Cmdable) *VersionManager {
	return &VersionManager{
		rdb: rdb,
		localCache: cache.New(
			cache.WithName("version_local_cache"),
			cache.WithLocal(local.NewTinyLFU(localCacheSize, localCacheTTL)),
			cache.WithErrNotFound(errVersionNotFound),
		),
	}
}

func (vm *VersionManager) GetVersion(ctx context.Context, entity string, id string) (int64, error) {
	versionKey := vm.versionKey(entity, id)

	var version int64
	err := vm.localCache.Once(ctx, versionKey, cache.Value(&version), cache.TTL(localCacheTTL),
		cache.Do(func(ctx context.Context) (any, error) {
			v, err := vm.rdb.Get(ctx, versionKey).Int64()
			if errors.Is(err, redis.Nil) {
				_, setErr := vm.rdb.SetNX(ctx, versionKey, int64(1), versionKeyTTL).Result()
				if setErr != nil {
					return nil, fmt.Errorf("failed to initialize version: %w", setErr)
				}
				return int64(1), nil
			}
			if err != nil {
				return nil, err
			}
			return v, nil
		}),
	)
	if err != nil {
		if errors.Is(err, errVersionNotFound) {
			return 1, nil
		}
		return 1, err
	}

	if version == 0 {
		return 1, nil
	}
	return version, nil
}

func (vm *VersionManager) GetMultiple(ctx context.Context, entities map[string][]string) (map[string]map[string]int64, error) {
	result := make(map[string]map[string]int64)

	for entity, ids := range entities {
		result[entity] = make(map[string]int64)
		for _, id := range ids {
			version, err := vm.GetVersion(ctx, entity, id)
			if err != nil {
				version = 1
			}
			result[entity][id] = version
		}
	}

	return result, nil
}

func (vm *VersionManager) GetEntityVersions(ctx context.Context, entity string) (globalVersion int64, err error) {
	return vm.GetVersion(ctx, entity, entityGlobalSuffix)
}

func (vm *VersionManager) Touch(ctx context.Context, entity string, id string) error {
	versionKey := vm.versionKey(entity, id)

	pipe := vm.rdb.Pipeline()
	pipe.Incr(ctx, versionKey)
	pipe.Expire(ctx, versionKey, versionKeyTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to touch version: %w", err)
	}

	_ = vm.localCache.Delete(ctx, versionKey)
	return nil
}

func (vm *VersionManager) TouchMultiple(ctx context.Context, entities map[string][]string) error {
	if len(entities) == 0 {
		return nil
	}

	pipe := vm.rdb.Pipeline()
	localKeys := make([]string, 0, len(entities)*2)

	for entity, ids := range entities {
		for _, id := range ids {
			versionKey := vm.versionKey(entity, id)
			pipe.Incr(ctx, versionKey)
			pipe.Expire(ctx, versionKey, versionKeyTTL)
			localKeys = append(localKeys, versionKey)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to touch multiple versions: %w", err)
	}

	for _, key := range localKeys {
		_ = vm.localCache.Delete(ctx, key)
	}

	return nil
}

func (vm *VersionManager) TouchEntity(ctx context.Context, entity string) error {
	return vm.Touch(ctx, entity, entityGlobalSuffix)
}

func (vm *VersionManager) TouchEntityMultiple(ctx context.Context, entities []string) error {
	if len(entities) == 0 {
		return nil
	}

	entityMap := make(map[string][]string)
	for _, entity := range entities {
		entityMap[entity] = []string{entityGlobalSuffix}
	}

	return vm.TouchMultiple(ctx, entityMap)
}

func (vm *VersionManager) versionKey(entity, id string) string {
	return fmt.Sprintf("%s:%s:%s", versionPrefix, entity, id)
}