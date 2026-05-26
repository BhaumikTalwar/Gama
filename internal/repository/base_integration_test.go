//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/BhaumikTalwar/Gama/internal/caching"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/BhaumikTalwar/Gama/internal/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBaseRepoTest(t *testing.T) (*BaseRepo, *Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	return repos.Base, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func TestBaseRepo_ExecTx_Success(t *testing.T) {
	_, repos, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "tx_success@example.com"

	err := repos.ExecTx(ctx, func(tx *Repositories) error {
		password := "hashed_password"
		_, err := tx.User.Create(ctx, db.CreateUserParams{
			Email:        &email,
			Username:     email,
			PhoneNumber:  "+919000000001",
			PasswordHash: password,
		})
		return err
	})
	require.NoError(t, err)

	user, err := repos.User.GetByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
}

func TestBaseRepo_ExecTx_RollbackOnError(t *testing.T) {
	_, repos, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "tx_rollback@example.com"

	err := repos.ExecTx(ctx, func(tx *Repositories) error {
		password := "hashed_password"
		_, err := tx.User.Create(ctx, db.CreateUserParams{
			Email:        &email,
			Username:     email,
			PhoneNumber:  "+919000000002",
			PasswordHash: password,
		})
		require.NoError(t, err)

		return assert.AnError
	})
	require.Error(t, err)

	_, err = repos.User.GetByEmail(ctx, email)
	assert.Error(t, err)
}

func TestBaseRepo_ExecTx_MultipleOperations(t *testing.T) {
	_, repos, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "tx_multi@example.com"
	roleName := "tx_multi_role"

	err := repos.ExecTx(ctx, func(tx *Repositories) error {
		password := "hashed_password"
		user, err := tx.User.Create(ctx, db.CreateUserParams{
			Email:        &email,
			Username:     email,
			PhoneNumber:  "+919000000003",
			PasswordHash: password,
		})
		if err != nil {
			return err
		}

		role, err := tx.RBAC.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
		if err != nil {
			return err
		}

		_, err = tx.RBAC.AssignRole(ctx, db.AssignRoleToUserParams{
			UserID: user.ID,
			RoleID: role.ID,
		})
		return err
	})
	require.NoError(t, err)

	user, err := repos.User.GetByEmail(ctx, email)
	require.NoError(t, err)

	roles, err := repos.RBAC.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, roleName, roles[0].Name)
}

func TestBaseRepo_ExecTx_RollbackMultiple(t *testing.T) {
	_, repos, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "tx_rollback_multi@example.com"
	roleName := "tx_rollback_role"

	err := repos.ExecTx(ctx, func(tx *Repositories) error {
		password := "hashed_password"
		user, err := tx.User.Create(ctx, db.CreateUserParams{
			Email:        &email,
			Username:     email,
			PhoneNumber:  "+919000000004",
			PasswordHash: password,
		})
		require.NoError(t, err)

		_, err = tx.RBAC.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
		require.NoError(t, err)

		_, err = tx.RBAC.AssignRole(ctx, db.AssignRoleToUserParams{
			UserID: user.ID,
			RoleID: 99999,
		})
		return err
	})
	require.Error(t, err)

	_, err = repos.User.GetByEmail(ctx, email)
	assert.Error(t, err)
}

func TestBaseRepo_TouchEntity(t *testing.T) {
	base, repos, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	err := base.TouchEntity(ctx, "user")
	require.NoError(t, err)

	err = base.TouchEntity(ctx, "user")
	require.NoError(t, err)

	_ = user
}

func TestBaseRepo_TouchEntityMultiple(t *testing.T) {
	base, _, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	err := base.TouchEntityMultiple(context.Background(), []string{"user", "role"})
	require.NoError(t, err)
}

func TestBaseRepo_GetVersionsForEntities(t *testing.T) {
	base, _, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	versions, err := base.GetVersionsForEntities(context.Background(), []string{"user", "role"})
	require.NoError(t, err)
	assert.NotNil(t, versions)
	assert.Contains(t, versions, "user")
	assert.Contains(t, versions, "role")
}

func TestBaseRepo_WithTxDB(t *testing.T) {
	base, repos, cleanup := setupBaseRepoTest(t)
	defer cleanup()

	txBase := base.WithTxDB(nil)
	assert.True(t, txBase.inTx)

	txBase2 := repos.RBAC.BaseRepo.WithTxDB(nil)
	assert.True(t, txBase2.inTx)
}
