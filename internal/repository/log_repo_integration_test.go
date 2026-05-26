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

func setupLogRepoTest(t *testing.T) (*LogRepo, *Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	return repos.Log, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func TestLogRepo_CreateAndGetLogs(t *testing.T) {
	repo, repos, cleanup := setupLogRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	desc := "User logged in"

	err := repo.CreateUserLog(ctx, db.CreateUserLogParams{
		UserID:      user.ID,
		Action:      "login",
		Description: &desc,
		IsSuccess:   true,
	})
	require.NoError(t, err)

	logs, err := repo.GetUserLogs(ctx, db.GetUserLogsParams{
		UserID:    user.ID,
		Limitval:  10,
		Offsetval: 0,
	})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "login", logs[0].Action)
	assert.True(t, logs[0].IsSuccess)
}

func TestLogRepo_CreateMultipleLogs(t *testing.T) {
	repo, repos, cleanup := setupLogRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	actions := []string{"login", "logout", "password_change", "profile_update"}
	for _, action := range actions {
		err := repo.CreateUserLog(ctx, db.CreateUserLogParams{
			UserID:    user.ID,
			Action:    action,
			IsSuccess: true,
		})
		require.NoError(t, err)
	}

	logs, err := repo.GetUserLogs(ctx, db.GetUserLogsParams{
		UserID:    user.ID,
		Limitval:  10,
		Offsetval: 0,
	})
	require.NoError(t, err)
	assert.Len(t, logs, 4)
}

func TestLogRepo_GetLogsPagination(t *testing.T) {
	repo, repos, cleanup := setupLogRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	for range 5 {
		err := repo.CreateUserLog(ctx, db.CreateUserLogParams{
			UserID:    user.ID,
			Action:    "test_action",
			IsSuccess: true,
		})
		require.NoError(t, err)
	}

	logs, err := repo.GetUserLogs(ctx, db.GetUserLogsParams{
		UserID:    user.ID,
		Limitval:  3,
		Offsetval: 0,
	})
	require.NoError(t, err)
	assert.Len(t, logs, 3)

	logs, err = repo.GetUserLogs(ctx, db.GetUserLogsParams{
		UserID:    user.ID,
		Limitval:  3,
		Offsetval: 3,
	})
	require.NoError(t, err)
	assert.Len(t, logs, 2)
}
