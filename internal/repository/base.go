package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/BhaumikTalwar/Gama/internal/caching"
	"github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
)

type BaseRepo struct {
	db         *db.Queries
	cache      caching.CacheService
	versionMgr *caching.VersionManager
	keyGen     *caching.KeyGen
	namespace  string
	inTx       bool
}

func NewBaseRepo(
	db *db.Queries,
	cache caching.CacheService,
	versionMgr *caching.VersionManager,
	namespace string,
) *BaseRepo {
	return &BaseRepo{
		db:         db,
		cache:      cache,
		versionMgr: versionMgr,
		keyGen:     caching.NewKeyGen(namespace),
		namespace:  namespace,
		inTx:       false,
	}
}

func (b *BaseRepo) WithTxDB(q *db.Queries) *BaseRepo {
	return &BaseRepo{
		db:         q,
		cache:      b.cache,
		versionMgr: b.versionMgr,
		keyGen:     b.keyGen,
		namespace:  b.namespace,
		inTx:       true,
	}
}

func Fetch[T any](
	ctx context.Context,
	repo *BaseRepo,
	key string,
	ttl time.Duration,
	loader func() (T, error),
) (T, error) {
	if repo.inTx {
		return loader()
	}

	val, err := repo.cache.Fetch(ctx, key, ttl, func() (any, error) {
		return loader()
	})
	if err != nil {
		var zero T
		return zero, fmt.Errorf("cache fetch failed: %w", err)
	}
	return val.(T), nil
}

func (b *BaseRepo) TouchEntity(ctx context.Context, entity string) error {
	if b.versionMgr == nil {
		return nil
	}
	return b.versionMgr.TouchEntity(ctx, entity)
}

func (b *BaseRepo) TouchEntityMultiple(ctx context.Context, entities []string) error {
	if b.versionMgr == nil {
		return nil
	}
	return b.versionMgr.TouchEntityMultiple(ctx, entities)
}

func (b *BaseRepo) InvalidateCache(ctx context.Context, pattern string) error {
	if b.cache == nil {
		return nil
	}
	return b.cache.DeletePattern(ctx, pattern)
}

func (b *BaseRepo) GetVersionsForEntities(ctx context.Context, entities []string) (map[string]map[string]int64, error) {
	if b.versionMgr == nil {
		return nil, nil
	}

	entityMap := make(map[string][]string)
	for _, entity := range entities {
		entityMap[entity] = []string{"global"}
	}
	return b.versionMgr.GetMultiple(ctx, entityMap)
}
