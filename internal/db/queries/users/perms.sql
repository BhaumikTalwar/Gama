-- name: CreatePermission :one
INSERT INTO permissions (name, resource, action, description)
VALUES (sqlc.arg(name), sqlc.arg(resource), sqlc.arg(action), sqlc.arg(description))
RETURNING *;

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = sqlc.arg(id);

-- name: GetPermissionByName :one
SELECT * FROM permissions WHERE name = sqlc.arg(name);

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY resource, action;

-- name: ListPermissionsByResource :many
SELECT * FROM permissions
WHERE resource = sqlc.arg(resource)
ORDER BY action;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = sqlc.arg(id);
