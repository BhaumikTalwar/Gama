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

func setupMFATest(t *testing.T) (*MFARepo, *Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	return repos.MFA, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func TestMFARepo_UpsertAndGetSettings_TOTP(t *testing.T) {
	repo, repos, cleanup := setupMFATest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	secret := "encrypted_secret_key_123"

	settings, err := repo.UpsertSettings(ctx, db.UpsertMFASettingsParams{
		UserID:    user.ID,
		SecretKey: &secret,
		Method:    db.MfaTypeTotp,
		Enabled:   false,
	})
	require.NoError(t, err)
	assert.Equal(t, db.MfaTypeTotp, settings.Method)
	assert.False(t, settings.Enabled)

	fetched, err := repo.GetSettings(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, db.MfaTypeTotp, fetched.UserMfaSetting.Method)
	assert.Equal(t, user.ID, fetched.UserMfaSetting.UserID)
}

func TestMFARepo_UpsertAndGetSettings_SMS(t *testing.T) {
	repo, repos, cleanup := setupMFATest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	phone := "+919000000010"

	settings, err := repo.UpsertSettings(ctx, db.UpsertMFASettingsParams{
		UserID:      user.ID,
		Method:      db.MfaTypeSms,
		PhoneNumber: &phone,
		Enabled:     false,
	})
	require.NoError(t, err)
	assert.Equal(t, db.MfaTypeSms, settings.Method)
	assert.Equal(t, phone, *settings.PhoneNumber)
}

func TestMFARepo_EnableAndDisable(t *testing.T) {
	repo, repos, cleanup := setupMFATest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	secret := "test_secret"

	_, err := repo.UpsertSettings(ctx, db.UpsertMFASettingsParams{
		UserID:    user.ID,
		SecretKey: &secret,
		Method:    db.MfaTypeTotp,
		Enabled:   false,
	})
	require.NoError(t, err)

	err = repo.SetEnabled(ctx, db.EnableMFAParams{
		UserID:      user.ID,
		BackupCodes: []string{"code1", "code2", "code3"},
	})
	require.NoError(t, err)

	fetched, err := repo.GetSettings(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, fetched.UserMfaSetting.Enabled)

	err = repo.Disable(ctx, user.ID)
	require.NoError(t, err)

	fetched, err = repo.GetSettings(ctx, user.ID)
	require.Error(t, err)
}

func TestMFARepo_GetSettings_NotFound(t *testing.T) {
	repo, repos, cleanup := setupMFATest(t)
	defer cleanup()

	user := createTestUser(t, repos)
	_, err := repo.GetSettings(context.Background(), user.ID)
	assert.Error(t, err)
}

func TestMFARepo_Upsert_UpdateSecret(t *testing.T) {
	repo, repos, cleanup := setupMFATest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	oldSecret := "old_secret"
	newSecret := "new_secret"

	_, err := repo.UpsertSettings(ctx, db.UpsertMFASettingsParams{
		UserID:    user.ID,
		SecretKey: &oldSecret,
		Method:    db.MfaTypeTotp,
		Enabled:   false,
	})
	require.NoError(t, err)

	_, err = repo.UpsertSettings(ctx, db.UpsertMFASettingsParams{
		UserID:    user.ID,
		SecretKey: &newSecret,
		Method:    db.MfaTypeTotp,
		Enabled:   false,
	})
	require.NoError(t, err)

	fetched, err := repo.GetSettings(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, newSecret, *fetched.UserMfaSetting.SecretKey)
}
