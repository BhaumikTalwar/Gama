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
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRBACTest(t *testing.T) (*RBACRepo, *Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	return repos.RBAC, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func uniqueName(prefix string) string {
	return prefix + "_" + uuid.New().String()[:8]
}

func TestRBAC_CreateAndGetRole(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	desc := "test role description"

	role, err := repo.CreateRole(ctx, db.CreateRoleParams{
		Name:        roleName,
		Description: &desc,
	})
	require.NoError(t, err)
	assert.Equal(t, roleName, role.Name)
	assert.True(t, role.ID > 0)

	fetched, err := repo.GetRoleByName(ctx, roleName)
	require.NoError(t, err)
	assert.Equal(t, role.ID, fetched.ID)
	assert.Equal(t, desc, *fetched.Description)
}

func TestRBAC_UpdateRole(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)

	newDesc := "updated description"
	updated, err := repo.UpdateRole(ctx, db.UpdateRoleParams{
		ID:          role.ID,
		Description: &newDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, newDesc, *updated.Description)
}

func TestRBAC_DeleteRole(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)

	err = repo.DeleteRole(ctx, role.ID)
	require.NoError(t, err)

	_, err = repo.GetRoleByName(ctx, roleName)
	assert.Error(t, err)
}

func TestRBAC_ListRoles(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	names := []string{uniqueName("role"), uniqueName("role"), uniqueName("role")}
	for _, name := range names {
		_, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: name})
		require.NoError(t, err)
	}

	roles, err := repo.ListRoles(ctx)
	require.NoError(t, err)
	roleNames := make(map[string]bool)
	for _, r := range roles {
		roleNames[r.Name] = true
	}
	for _, n := range names {
		assert.True(t, roleNames[n])
	}
}

func TestRBAC_CreateAndListPermissions(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	permName := uniqueName("perm")
	perm, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name:     permName,
		Resource: "test",
		Action:   "read",
	})
	require.NoError(t, err)
	assert.Equal(t, permName, perm.Name)

	permName2 := uniqueName("perm")
	perm2, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name:     permName2,
		Resource: "test",
		Action:   "write",
	})
	require.NoError(t, err)
	assert.NotEqual(t, perm.ID, perm2.ID)

	perms, err := repo.ListPermissions(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(perms), 2)
}

func TestRBAC_DeletePermission(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	permName := uniqueName("perm")
	perm, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name:     permName,
		Resource: "test",
		Action:   "delete",
	})
	require.NoError(t, err)

	err = repo.DeletePermission(ctx, perm.ID)
	require.NoError(t, err)
}

func TestRBAC_AssignRoleToUser(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	user := createTestUser(t, repos)

	ur, err := repo.AssignRole(ctx, db.AssignRoleToUserParams{
		UserID: user.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, user.ID, ur.UserID)
	assert.Equal(t, role.ID, ur.RoleID)
}

func TestRBAC_RevokeRoleFromUser(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	user := createTestUser(t, repos)

	_, err = repo.AssignRole(ctx, db.AssignRoleToUserParams{
		UserID: user.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)

	err = repo.RevokeRoleFromUser(ctx, db.RevokeRoleFromUserParams{
		UserID: user.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	for _, r := range roles {
		assert.NotEqual(t, role.ID, r.ID)
	}
}

func TestRBAC_GetUserRoles(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	role1, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: uniqueName("role")})
	require.NoError(t, err)
	role2, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: uniqueName("role")})
	require.NoError(t, err)
	user := createTestUser(t, repos)

	_, err = repo.AssignRole(ctx, db.AssignRoleToUserParams{UserID: user.ID, RoleID: role1.ID})
	require.NoError(t, err)
	_, err = repo.AssignRole(ctx, db.AssignRoleToUserParams{UserID: user.ID, RoleID: role2.ID})
	require.NoError(t, err)

	roles, err := repo.GetUserRoles(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRBAC_CheckPermission(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	permName := uniqueName("perm")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	perm, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name: permName, Resource: "test", Action: "read",
	})
	require.NoError(t, err)
	user := createTestUser(t, repos)

	_, err = repo.AssignRole(ctx, db.AssignRoleToUserParams{UserID: user.ID, RoleID: role.ID})
	require.NoError(t, err)
	err = repo.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{
		RoleID: role.ID, PermissionID: perm.ID,
	})
	require.NoError(t, err)

	hasPerm, err := repo.CheckPermission(ctx, db.CheckUserHasPermissionParams{
		UserID:         user.ID,
		PermissionName: permName,
	})
	require.NoError(t, err)
	assert.True(t, hasPerm)

	hasPerm, err = repo.CheckPermission(ctx, db.CheckUserHasPermissionParams{
		UserID:         user.ID,
		PermissionName: "nonexistent",
	})
	require.NoError(t, err)
	assert.False(t, hasPerm)
}

func TestRBAC_AssignPermissionToRole(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	permName := uniqueName("perm")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	perm, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name: permName, Resource: "test", Action: "action",
	})
	require.NoError(t, err)

	err = repo.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{
		RoleID: role.ID, PermissionID: perm.ID,
	})
	require.NoError(t, err)

	perms, err := repo.GetRolePermissions(ctx, role.ID)
	require.NoError(t, err)
	assert.Len(t, perms, 1)
	assert.Equal(t, permName, perms[0].Name)
}

func TestRBAC_RevokePermissionFromRole(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	permName := uniqueName("perm")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	perm, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name: permName, Resource: "test", Action: "action",
	})
	require.NoError(t, err)

	err = repo.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{
		RoleID: role.ID, PermissionID: perm.ID,
	})
	require.NoError(t, err)

	err = repo.RevokePermissionFromRole(ctx, db.RevokePermissionFromRoleParams{
		RoleID: role.ID, PermissionID: perm.ID,
	})
	require.NoError(t, err)

	perms, err := repo.GetRolePermissions(ctx, role.ID)
	require.NoError(t, err)
	assert.Empty(t, perms)
}

func TestRBAC_GetRolePermissions(t *testing.T) {
	repo, _, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	permName1 := uniqueName("perm")
	permName2 := uniqueName("perm")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	perm1, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name: permName1, Resource: "test", Action: "one",
	})
	require.NoError(t, err)
	perm2, err := repo.CreatePermission(ctx, db.CreatePermissionParams{
		Name: permName2, Resource: "test", Action: "two",
	})
	require.NoError(t, err)

	err = repo.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{RoleID: role.ID, PermissionID: perm1.ID})
	require.NoError(t, err)
	err = repo.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{RoleID: role.ID, PermissionID: perm2.ID})
	require.NoError(t, err)

	perms, err := repo.GetRolePermissions(ctx, role.ID)
	require.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestRBAC_CountUsersByRole(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)

	for range 3 {
		u := createTestUser(t, repos)
		_, err := repo.AssignRole(ctx, db.AssignRoleToUserParams{UserID: u.ID, RoleID: role.ID})
		require.NoError(t, err)
	}

	counts, err := repo.CountUsersByRole(ctx)
	require.NoError(t, err)

	found := false
	for _, c := range counts {
		if c.Name == roleName {
			assert.Equal(t, int64(3), c.UserCount)
			found = true
		}
	}
	assert.True(t, found)
}

func TestRBAC_ListUsersWithRole(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)

	u := createTestUser(t, repos)
	_, err = repo.AssignRole(ctx, db.AssignRoleToUserParams{UserID: u.ID, RoleID: role.ID})
	require.NoError(t, err)

	users, err := repo.ListUsersWithRole(ctx, role.ID)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestRBAC_GetUsersByRoles(t *testing.T) {
	repo, repos, cleanup := setupRBACTest(t)
	defer cleanup()

	ctx := context.Background()
	roleName := uniqueName("role")
	role, err := repo.CreateRole(ctx, db.CreateRoleParams{Name: roleName})
	require.NoError(t, err)
	u := createTestUser(t, repos)
	_, err = repo.AssignRole(ctx, db.AssignRoleToUserParams{UserID: u.ID, RoleID: role.ID})
	require.NoError(t, err)

	users, err := repo.GetUsersByRoles(ctx, []string{roleName})
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

var testPhoneCounter int

func createTestUser(t *testing.T, repos *Repositories) *db.User {
	t.Helper()
	testPhoneCounter++
	email := uuid.New().String() + "@test.com"
	password, _ := utils.HashPassword("testpass123")
	phone := fmt.Sprintf("+9190000000%02d", testPhoneCounter%100)
	fn := "Test"
	user, err := repos.User.Create(context.Background(), db.CreateUserParams{
		Email:        &email,
		Username:     email,
		PhoneNumber:  phone,
		PasswordHash: password,
		FirstName:    &fn,
	})
	require.NoError(t, err)
	return user
}
