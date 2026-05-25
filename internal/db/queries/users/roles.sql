-- name: CreateRole :one
INSERT INTO roles (name, description, is_system)
VALUES (sqlc.arg(name), sqlc.arg(description), sqlc.arg(is_system))
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = sqlc.arg(id);

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = sqlc.arg(name);

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: UpdateRole :one
UPDATE roles
SET
  description = COALESCE(sqlc.narg('description'), description)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = sqlc.arg(id) AND is_system = FALSE;
