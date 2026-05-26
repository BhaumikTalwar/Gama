//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/BhaumikTalwar/Gama/internal/caching"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/BhaumikTalwar/Gama/internal/testutil"
	"github.com/BhaumikTalwar/Gama/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserRepoTest(t *testing.T) (*UserRepo, *Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	return repos.User, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func TestUserRepo_CreateAndGetByID(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"
	password, _ := utils.HashPassword("password123")
	firstName := "John"
	lastName := "Doe"

	user, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "+919000000001",
		PasswordHash: password,
		FirstName:    &firstName,
		LastName:     &lastName,
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, email, user.Email)

	fetched, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, fetched.ID)
	assert.Equal(t, email, fetched.Email)
	assert.Equal(t, firstName, *fetched.FirstName)
	assert.Equal(t, lastName, *fetched.LastName)
}

func TestUserRepo_GetByEmail(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "findme@example.com"
	password, _ := utils.HashPassword("pass123")

	created, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "9111111111",
		PasswordHash: password,
	})
	require.NoError(t, err)

	fetched, err := repo.GetByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestUserRepo_GetByPhone(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	phone := "9999999999"
	email := "phone@example.com"
	password, _ := utils.HashPassword("pass123")

	created, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  phone,
		PasswordHash: password,
	})
	require.NoError(t, err)

	fetched, err := repo.GetByPhone(ctx, phone)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
}

func TestUserRepo_GetByUsername(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	username := "uniqueuser"
	email := "uname@example.com"
	password, _ := utils.HashPassword("pass123")

	created, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     username,
		PhoneNumber:  "9222222222",
		PasswordHash: password,
	})
	require.NoError(t, err)

	fetched, err := repo.GetByUsername(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
}

func TestUserRepo_DuplicateEmail(t *testing.T) {
	repo, repos, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "dupe@example.com"
	password, _ := utils.HashPassword("pass123")

	_, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "9333333333",
		PasswordHash: password,
	})
	require.NoError(t, err)

	_, err = repos.User.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email + "-alt",
		PhoneNumber:  "9444444444",
		PasswordHash: password,
	})
	assert.Error(t, err)
}

func TestUserRepo_Update(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "update@example.com"
	password, _ := utils.HashPassword("pass123")
	newFirstName := "Updated"
	newLastName := "User"

	user, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "9555555555",
		PasswordHash: password,
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, db.UpdateUserParams{
		ID:        user.ID,
		FirstName: &newFirstName,
		LastName:  &newLastName,
	})
	require.NoError(t, err)
	assert.Equal(t, newFirstName, *updated.FirstName)
	assert.Equal(t, newLastName, *updated.LastName)
}

func TestUserRepo_UpdatePassword(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "pwd@example.com"
	oldPassword, _ := utils.HashPassword("oldpass")
	newPassword, _ := utils.HashPassword("newpass")

	user, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "9666666666",
		PasswordHash: oldPassword,
	})
	require.NoError(t, err)

	err = repo.UpdatePassword(ctx, db.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: newPassword,
	})
	require.NoError(t, err)

	fetched, _ := repo.GetByID(ctx, user.ID)
	assert.True(t, utils.CheckPasswordHash("newpass", fetched.PasswordHash))
	assert.False(t, utils.CheckPasswordHash("oldpass", fetched.PasswordHash))
}

func TestUserRepo_SoftDelete(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "delete@example.com"
	password, _ := utils.HashPassword("pass123")

	user, err := repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "9777777777",
		PasswordHash: password,
	})
	require.NoError(t, err)

	err = repo.SoftDeleteUser(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
}

func TestUserRepo_ListUsers(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	password, _ := utils.HashPassword("pass123")

	for i := range 5 {
		e := fmt.Sprintf("list%d@example.com", i)
		_, err := repo.Create(ctx, db.CreateUserParams{
			Email:        &e,
			Username:     e,
			PhoneNumber:  fmt.Sprintf("9%09d", i),
			PasswordHash: password,
		})
		require.NoError(t, err)
	}

	users, err := repo.ListUsers(ctx, db.ListUsersParams{Limitval: 10, Offsetval: 0})
	require.NoError(t, err)
	assert.Len(t, users, 5)
}

func TestUserRepo_SearchUsers(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	password, _ := utils.HashPassword("pass123")

	emails := []string{"alice@example.com", "bob@example.com", "charlie@example.com"}
	for i, e := range emails {
		_, err := repo.Create(ctx, db.CreateUserParams{
			Email:        &e,
			Username:     e,
			PhoneNumber:  fmt.Sprintf("9%09d", i),
			PasswordHash: password,
		})
		require.NoError(t, err)
	}

	search := "alice"
	results, err := repo.SearchUsers(ctx, db.SearchUsersParams{
		SearchQuery: &search,
		Limitval:    10,
		Offsetval:   0,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Email, "alice")
}

func TestUserRepo_CountUsers(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	password, _ := utils.HashPassword("pass123")

	count, err := repo.CountUsers(ctx)
	require.NoError(t, err)
	initial := count

	email := "count@example.com"
	_, err = repo.Create(ctx, db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  "9888888888",
		PasswordHash: password,
	})
	require.NoError(t, err)

	count, err = repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, initial+1, count)
}

func TestUserRepo_GetUsersByIDs(t *testing.T) {
	repo, _, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	password, _ := utils.HashPassword("pass123")

	var ids []uuid.UUID
	for i := range 3 {
		e := fmt.Sprintf("ids%d@example.com", i)
		u, err := repo.Create(ctx, db.CreateUserParams{
			Email:        &e,
			Username:     e,
			PhoneNumber:  fmt.Sprintf("9%09d", i),
			PasswordHash: password,
		})
		require.NoError(t, err)
		ids = append(ids, u.ID)
	}

	users, err := repo.GetUsersByIDs(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestUserRepo_Transaction_CreateUserWithRole(t *testing.T) {
	repo, repos, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	role, err := repos.RBAC.CreateRole(ctx, db.CreateRoleParams{
		Name: "test_user",
	})
	require.NoError(t, err)

	var createdUserID uuid.UUID
	err = repos.ExecTx(ctx, func(tx *Repositories) error {
		email := "txuser@example.com"
		password, _ := utils.HashPassword("pass123")
		fn := "Tx"

		user, err := tx.User.Create(ctx, db.CreateUserParams{
			Email:        &email,
			Username:     email,
			PhoneNumber:  "9999999990",
			PasswordHash: password,
			FirstName:    &fn,
		})
		if err != nil {
			return err
		}
		createdUserID = user.ID

		_, err = tx.RBAC.AssignRole(ctx, db.AssignRoleToUserParams{
			UserID: user.ID,
			RoleID: role.ID,
		})
		return err
	})
	require.NoError(t, err)

	fetched, err := repo.GetByID(ctx, createdUserID)
	require.NoError(t, err)
	assert.Equal(t, "Tx", *fetched.FirstName)
}
