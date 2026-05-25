package repository

import (
	"context"

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
	return r.db.GetRolePermissions(ctx, roleID)
}

func (r *RBACRepo) CheckPermission(ctx context.Context, arg db.CheckUserHasPermissionParams) (bool, error) {
	return r.db.CheckUserHasPermission(ctx, arg)
}

func (r *RBACRepo) AssignRole(ctx context.Context, arg db.AssignRoleToUserParams) (*db.UserRole, error) {
	role, err := r.db.AssignRoleToUser(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RBACRepo) RevokeRoleFromUser(ctx context.Context, arg db.RevokeRoleFromUserParams) error {
	err := r.db.RevokeRoleFromUser(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}

func (r *RBACRepo) CreateRole(ctx context.Context, arg db.CreateRoleParams) (*db.Role, error) {
	role, err := r.db.CreateRole(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RBACRepo) UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (*db.Role, error) {
	role, err := r.db.UpdateRole(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RBACRepo) DeleteRole(ctx context.Context, id int32) error {
	err := r.db.DeleteRole(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *RBACRepo) ListRoles(ctx context.Context) ([]db.Role, error) {
	return r.db.ListRoles(ctx)
}

func (r *RBACRepo) CreatePermission(ctx context.Context, arg db.CreatePermissionParams) (*db.Permission, error) {
	perm, err := r.db.CreatePermission(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &perm, nil
}

func (r *RBACRepo) DeletePermission(ctx context.Context, id int32) error {
	err := r.db.DeletePermission(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *RBACRepo) ListPermissions(ctx context.Context) ([]db.Permission, error) {
	return r.db.ListPermissions(ctx)
}

func (r *RBACRepo) AssignPermissionToRole(ctx context.Context, arg db.AssignPermissionToRoleParams) error {
	err := r.db.AssignPermissionToRole(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}

func (r *RBACRepo) RevokePermissionFromRole(ctx context.Context, arg db.RevokePermissionFromRoleParams) error {
	err := r.db.RevokePermissionFromRole(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}
