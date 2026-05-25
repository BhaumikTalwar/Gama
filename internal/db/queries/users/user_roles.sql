-- name: AssignRoleToUser :one
INSERT INTO user_roles (
  user_id,
  role_id,
  scope_type,
  scope_id,
  assigned_by,
  expires_at
) VALUES (
  sqlc.arg(user_id), sqlc.arg(role_id), sqlc.arg(scope_type), sqlc.arg(scope_id), sqlc.arg(assigned_by), sqlc.arg(expires_at)
) RETURNING *;

-- name: RevokeRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = sqlc.arg(user_id) 
  AND role_id = sqlc.arg(role_id)
  AND COALESCE(scope_type, '') = COALESCE(sqlc.arg(scope_type), '')
  AND COALESCE(scope_id, '') = COALESCE(sqlc.arg(scope_id), '');

-- name: GetUserRoles :many
SELECT r.*, ur.scope_type, ur.scope_id, ur.assigned_at, ur.expires_at
FROM roles r
INNER JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id)
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

-- name: GetUserRolesByScope :many
SELECT r.*, ur.scope_type, ur.scope_id
FROM roles r
INNER JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id)
  AND ur.scope_type = sqlc.arg(scope_type)
  AND ur.scope_id = sqlc.arg(scope_id)
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

-- name: GetUserGlobalRoles :many
SELECT r.* FROM roles r
INNER JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id)
  AND ur.scope_type IS NULL
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

-- name: GetUserPermissions :many
SELECT DISTINCT p.*
FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id)
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
ORDER BY p.resource, p.action;

-- name: GetUserPermissionsByScope :many
SELECT DISTINCT p.*
FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id)
  AND (
    (ur.scope_type = sqlc.arg(scope_type) AND ur.scope_id = sqlc.arg(scope_id)) OR
    ur.scope_type IS NULL
  )
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
ORDER BY p.resource, p.action;

-- name: CheckUserHasRole :one
SELECT EXISTS(
  SELECT 1
  FROM user_roles ur
  INNER JOIN roles r ON ur.role_id = r.id
  WHERE ur.user_id = sqlc.arg(user_id)
    AND r.name = sqlc.arg(role_name)
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
);

-- name: CheckUserHasRoleInScope :one
SELECT EXISTS(
  SELECT 1
  FROM user_roles ur
  INNER JOIN roles r ON ur.role_id = r.id
  WHERE ur.user_id = sqlc.arg(user_id)
    AND r.name = sqlc.arg(role_name)
    AND (
      (ur.scope_type = sqlc.arg(scope_type) AND ur.scope_id = sqlc.arg(scope_id)) OR
      ur.scope_type IS NULL
    )
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
);

-- name: CountUsersByRole :many
SELECT 
  r.id,
  r.name,
  COUNT(DISTINCT ur.user_id) as user_count
FROM roles r
LEFT JOIN user_roles ur ON r.id = ur.role_id 
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
GROUP BY r.id, r.name
ORDER BY user_count DESC, r.name;

-- name: CheckUserHasPermission :one
SELECT EXISTS(
  SELECT 1
  FROM permissions p
  INNER JOIN role_permissions rp ON p.id = rp.permission_id
  INNER JOIN user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = sqlc.arg(user_id)
    AND p.name = sqlc.arg(permission_name)
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
);

-- name: CheckUserHasPermissionInScope :one
SELECT EXISTS(
  SELECT 1
  FROM permissions p
  INNER JOIN role_permissions rp ON p.id = rp.permission_id
  INNER JOIN user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = sqlc.arg(user_id)
    AND p.name = sqlc.arg(permission_name)
    AND (
      (ur.scope_type = sqlc.arg(scope_type) AND ur.scope_id = sqlc.arg(scope_id)) OR
      ur.scope_type IS NULL
    )
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
);

-- name: ListUsersWithRole :many
SELECT u.* FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role_id = sqlc.arg(role_id)
  AND u.deleted_at IS NULL
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
ORDER BY u.username;

-- name: GetUsersByRoles :many
SELECT DISTINCT u.* FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
INNER JOIN roles r ON ur.role_id = r.id
WHERE r.name = ANY(sqlc.arg(role_names)::text[])
  AND u.deleted_at IS NULL
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
ORDER BY u.username;

-- name: BulkRevokeUserRoles :exec
DELETE FROM user_roles
WHERE id = ANY(sqlc.arg(user_role_ids)::uuid[]);
