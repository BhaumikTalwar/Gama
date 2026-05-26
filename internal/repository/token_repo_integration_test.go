//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/internal/caching"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/BhaumikTalwar/Gama/internal/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTokenRepoTest(t *testing.T) (*TokenRepo, *Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	return repos.Token, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func TestTokenRepo_CreateAndGetRefreshToken(t *testing.T) {
	repo, repos, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	token, err := repo.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: "test_hash_123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)
	assert.NotEqual(t, token.ID, "")
	assert.False(t, token.Revoked)

	fetched, err := repo.GetRefreshToken(ctx, "test_hash_123")
	require.NoError(t, err)
	assert.Equal(t, token.ID, fetched.ID)
	assert.Equal(t, user.ID, fetched.UserID)
}

func TestTokenRepo_GetRefreshToken_NotFound(t *testing.T) {
	repo, _, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	_, err := repo.GetRefreshToken(context.Background(), "nonexistent_hash")
	assert.Error(t, err)
}

func TestTokenRepo_RevokeRefreshToken(t *testing.T) {
	repo, repos, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	reason := "test_revocation"

	token, err := repo.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: "revoke_test_hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	err = repo.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		ID:            token.ID,
		RevokedReason: &reason,
	})
	require.NoError(t, err)

	_, err = repo.GetRefreshToken(ctx, "revoke_test_hash")
	assert.Error(t, err)
}

func TestTokenRepo_RevokeAllUserTokens(t *testing.T) {
	repo, repos, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	for i := range 3 {
		h := "bulk_revoke_hash_" + string(rune('0'+i))
		_, err := repo.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			UserID:    user.ID,
			TokenHash: h,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
	}

	err := repo.RevokeAllUserTokens(ctx, user.ID)
	require.NoError(t, err)
}

func TestTokenRepo_CreateAndGetVerificationToken(t *testing.T) {
	repo, repos, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)
	externalID := "ext_123"

	token, err := repo.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		UserID:     user.ID,
		TokenHash:  "verify_hash_123",
		TokenType:  db.TokenTypeEmailVerification,
		ExternalID: &externalID,
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	})
	require.NoError(t, err)
	assert.NotEqual(t, token.ID, "")

	fetched, err := repo.GetVerificationTokenByHash(ctx, db.GetVerificationTokenByHashParams{
		TokenHash: "verify_hash_123",
		TokenType: db.TokenTypeEmailVerification,
	})
	require.NoError(t, err)
	assert.Equal(t, token.ID, fetched.ID)

	fetchedByExt, err := repo.GetVerificationTokenByExternalID(ctx, db.GetVerificationTokenByExternalIDParams{
		UserID:     user.ID,
		TokenType:  db.TokenTypeEmailVerification,
		ExternalID: &externalID,
	})
	require.NoError(t, err)
	assert.Equal(t, token.ID, fetchedByExt.ID)
}

func TestTokenRepo_GetLatestVerificationToken(t *testing.T) {
	repo, repos, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	for i := range 3 {
		h := "latest_hash_" + string(rune('0'+i))
		e := "ext_" + string(rune('0'+i))
		_, err := repo.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
			UserID:     user.ID,
			TokenHash:  h,
			TokenType:  db.TokenTypeSmsOtp,
			ExternalID: &e,
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	latest, err := repo.GetLatestVerificationTokenForUser(ctx, db.GetLatestVerificationTokenForUserParams{
		UserID:    user.ID,
		TokenType: db.TokenTypeSmsOtp,
	})
	require.NoError(t, err)
	assert.Equal(t, "latest_hash_2", latest.TokenHash)
}

func TestTokenRepo_MarkTokenUsed(t *testing.T) {
	repo, repos, cleanup := setupTokenRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, repos)

	token, err := repo.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		UserID:    user.ID,
		TokenHash: "mark_used_hash",
		TokenType: db.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})
	require.NoError(t, err)

	err = repo.MarkTokenUsed(ctx, token.ID)
	require.NoError(t, err)
}
