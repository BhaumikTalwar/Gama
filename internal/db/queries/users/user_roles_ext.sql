-- name: RevokeUserRoleByID :exec
DELETE FROM user_roles WHERE id = sqlc.arg(id);

-- name: UpdateUserRole :one
UPDATE user_roles
SET expires_at = sqlc.narg('expires_at')
WHERE id = sqlc.arg('id')
RETURNING *;
