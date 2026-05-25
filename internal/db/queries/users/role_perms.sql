-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id, granted_by)
VALUES (sqlc.arg(role_id), sqlc.arg(permission_id), sqlc.arg(granted_by))
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: RevokePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = sqlc.arg(role_id) AND permission_id = sqlc.arg(permission_id);

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = sqlc.arg(role_id)
ORDER BY p.resource, p.action;

-- name: CheckRoleHasPermission :one
SELECT EXISTS(
  SELECT 1 FROM role_permissions
  WHERE role_id = sqlc.arg(role_id) AND permission_id = sqlc.arg(permission_id)
);
