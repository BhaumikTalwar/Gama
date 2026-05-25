package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/mgtv-tech/jetcache-go"
)

type BaseRepo struct {
	db    *db.Queries
	cache cache.Cache
	inTx  bool
}

func NewBaseRepo(
	db *db.Queries,
	cache cache.Cache,
) *BaseRepo {
	return &BaseRepo{
		db:    db,
		cache: cache,
		inTx:  false,
	}
}

func (b *BaseRepo) WithTxDB(q *db.Queries) *BaseRepo {
	return &BaseRepo{
		db:    q,
		cache: b.cache,
		inTx:  true,
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

	var result T
	err := repo.cache.Once(ctx, key, cache.Value(&result), cache.TTL(ttl),
		cache.Do(func(ctx context.Context) (any, error) {
			return loader()
		}))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("cache fetch failed: %w", err)
	}

	return result, nil
}
