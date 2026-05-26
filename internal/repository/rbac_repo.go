package repository

import (
	"context"
	"fmt"
	"time"

	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/google/uuid"
)

type RBACRepo struct {
	*BaseRepo
}

func NewRBACRepo(base *BaseRepo) *RBACRepo {
	return &RBACRepo{BaseRepo: base}
}

func (r *RBACRepo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]db.GetUserRolesRow, error) {
	return r.db.GetUserRoles(ctx, userID)
}

func (r *RBACRepo) GetRoleByName(ctx context.Context, name string) (db.Role, error) {
	return r.db.GetRoleByName(ctx, name)
}

func (r *RBACRepo) GetRolePermissions(ctx context.Context, roleID int32) ([]db.Permission, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"role"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, r.keyGen.WithParams("role_perms", "id", roleID))

	return Fetch(ctx, r.BaseRepo, key, 10*time.Minute, func() ([]db.Permission, error) {
		return r.db.GetRolePermissions(ctx, roleID)
	})
}

func (r *RBACRepo) CheckPermission(ctx context.Context, arg db.CheckUserHasPermissionParams) (bool, error) {
	return r.db.CheckUserHasPermission(ctx, arg)
}

func (r *RBACRepo) AssignRole(ctx context.Context, arg db.AssignRoleToUserParams) (*db.UserRole, error) {
	role, err := r.db.AssignRoleToUser(ctx, arg)
	if err != nil {
		return nil, err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("role", "global", "*"))

	return &role, nil
}

func (r *RBACRepo) RevokeRoleFromUser(ctx context.Context, arg db.RevokeRoleFromUserParams) error {
	err := r.db.RevokeRoleFromUser(ctx, arg)
	if err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("role", "global", "*"))

	return nil
}

func (r *RBACRepo) CreateRole(ctx context.Context, arg db.CreateRoleParams) (*db.Role, error) {
	role, err := r.db.CreateRole(ctx, arg)
	if err != nil {
		return nil, err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("role", "global", "list"))

	return &role, nil
}

func (r *RBACRepo) UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (*db.Role, error) {
	role, err := r.db.UpdateRole(ctx, arg)
	if err != nil {
		return nil, err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.WithParams(r.keyGen.Simple("role", fmt.Sprintf("%d", arg.ID), "detail")))
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("role", "global", "list"))

	return &role, nil
}

func (r *RBACRepo) DeleteRole(ctx context.Context, id int32) error {
	err := r.db.DeleteRole(ctx, id)
	if err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.WithParams(r.keyGen.Simple("role", fmt.Sprintf("%d", id), "detail")))
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("role", "global", "list"))
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("role", "global", "*"))

	return nil
}

func (r *RBACRepo) ListRoles(ctx context.Context) ([]db.Role, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"role"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, "list")

	return Fetch(ctx, r.BaseRepo, key, 10*time.Minute, func() ([]db.Role, error) {
		return r.db.ListRoles(ctx)
	})
}

func (r *RBACRepo) CreatePermission(ctx context.Context, arg db.CreatePermissionParams) (*db.Permission, error) {
	perm, err := r.db.CreatePermission(ctx, arg)
	if err != nil {
		return nil, err
	}

	_ = r.TouchEntity(ctx, "permission")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("permission", "global", "list"))

	return &perm, nil
}

func (r *RBACRepo) DeletePermission(ctx context.Context, id int32) error {
	err := r.db.DeletePermission(ctx, id)
	if err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "permission")
	_ = r.InvalidateCache(ctx, r.keyGen.WithParams(r.keyGen.Simple("permission", fmt.Sprintf("%d", id), "detail")))
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("permission", "global", "list"))

	return nil
}

func (r *RBACRepo) ListPermissions(ctx context.Context) ([]db.Permission, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"permission"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, "list")

	return Fetch(ctx, r.BaseRepo, key, 10*time.Minute, func() ([]db.Permission, error) {
		return r.db.ListPermissions(ctx)
	})
}

func (r *RBACRepo) AssignPermissionToRole(ctx context.Context, arg db.AssignPermissionToRoleParams) error {
	err := r.db.AssignPermissionToRole(ctx, arg)
	if err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.WithParams(r.keyGen.Simple("role", fmt.Sprintf("%d", arg.RoleID), "perms")))

	return nil
}

func (r *RBACRepo) RevokePermissionFromRole(ctx context.Context, arg db.RevokePermissionFromRoleParams) error {
	err := r.db.RevokePermissionFromRole(ctx, arg)
	if err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "role")
	_ = r.InvalidateCache(ctx, r.keyGen.WithParams(r.keyGen.Simple("role", fmt.Sprintf("%d", arg.RoleID), "perms")))

	return nil
}

func (r *RBACRepo) CountUsersByRole(ctx context.Context) ([]db.CountUsersByRoleRow, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"role"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, "count_by_role")

	return Fetch(ctx, r.BaseRepo, key, 5*time.Minute, func() ([]db.CountUsersByRoleRow, error) {
		return r.db.CountUsersByRole(ctx)
	})
}

func (r *RBACRepo) ListUsersWithRole(ctx context.Context, roleID int32) ([]db.User, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"user", "role"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, r.keyGen.WithParams("list_with_role", "roleID", roleID))

	return Fetch(ctx, r.BaseRepo, key, 5*time.Minute, func() ([]db.User, error) {
		return r.db.ListUsersWithRole(ctx, roleID)
	})
}

func (r *RBACRepo) GetUsersByRoles(ctx context.Context, roleNames []string) ([]db.User, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"user", "role"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, r.keyGen.WithParams("get_by_roles", "roles", roleNames))

	return Fetch(ctx, r.BaseRepo, key, 5*time.Minute, func() ([]db.User, error) {
		return r.db.GetUsersByRoles(ctx, roleNames)
	})
}
