package repository

import (
	"context"
	"fmt"

	"github.com/BhaumikTalwar/Gama/internal/caching"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	Base  *BaseRepo
	User  *UserRepo
	MFA   *MFARepo
	RBAC  *RBACRepo
	Token *TokenRepo
	Log   *LogRepo
	pool  *pgxpool.Pool
}

func NewRepositories(
	pool *pgxpool.Pool,
	base *BaseRepo,
) *Repositories {
	return &Repositories{
		pool:  pool,
		Base:  base,
		User:  NewUserRepo(base),
		MFA:   NewMfaRepo(base),
		RBAC:  NewRBACRepo(base),
		Token: NewTokenRepo(base),
		Log:   NewLogRepo(base),
	}
}

func SetupPostgresRepositories(
	pool *pgxpool.Pool,
	cache caching.CacheService,
	versionMgr *caching.VersionManager,
	namespace string,
) *Repositories {
	q := db.New(pool)
	base := NewBaseRepo(q, cache, versionMgr, namespace)
	return NewRepositories(pool, base)
}

func (r *Repositories) ExecTx(ctx context.Context, fn func(*Repositories) error, opts ...pgx.TxOptions) error {
	if r.pool == nil {
		return fmt.Errorf("transaction is already in progress or pool is not set")
	}

	txOptions := pgx.TxOptions{}
	if len(opts) > 0 {
		txOptions = opts[0]
	}

	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qTx := db.New(tx)
	txBase := r.Base.WithTxDB(qTx)
	txRepo := NewRepositories(nil, txBase)

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func RunInTx[T any](ctx context.Context, r *Repositories, fn func(*Repositories) (T, error)) (T, error) {
	var result T
	err := r.ExecTx(ctx, func(txRepo *Repositories) error {
		var innerErr error
		result, innerErr = fn(txRepo)
		return innerErr
	})

	return result, err
}