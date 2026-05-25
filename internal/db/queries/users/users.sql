-- name: CreateUser :one
INSERT INTO users (
  username,
  phone_number,
  email,
  password_hash,
  first_name,
  last_name
) VALUES (
  sqlc.arg(username), sqlc.arg(phone_number), sqlc.narg(email), sqlc.arg(password_hash), sqlc.narg(first_name), sqlc.narg(last_name)
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = sqlc.arg(email) AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = sqlc.arg(username) AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT sqlc.arg(limitVal) OFFSET sqlc.arg(offsetVal);

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET
  first_name = COALESCE(sqlc.narg('first_name'), first_name),
  last_name = COALESCE(sqlc.narg('last_name'), last_name),
  avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = sqlc.arg(password_hash)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = TRUE
WHERE id = sqlc.arg(id);

-- name: DisableUser :exec
UPDATE users
SET disabled = TRUE
WHERE id = sqlc.arg(id);

-- name: EnableUser :exec
UPDATE users
SET disabled = FALSE
WHERE id = sqlc.arg(id);

-- name: SoftDeleteUser :exec
UPDATE users
SET 
    deleted_at = NOW(),
    disabled = TRUE
WHERE id = sqlc.arg(id);

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = sqlc.arg(id);

-- name: SearchUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND (
    username ILIKE '%' || sqlc.arg(search_query) || '%' OR
    email ILIKE '%' || sqlc.arg(search_query) || '%' OR
    phone_number ILIKE '%' || sqlc.arg(search_query) || '%' OR
    first_name ILIKE '%' || sqlc.arg(search_query) || '%' OR
    last_name ILIKE '%' || sqlc.arg(search_query) || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(limitVal) OFFSET sqlc.arg(offsetVal);

-- name: CountUsersBySearch :one
SELECT COUNT(*)::bigint FROM users
WHERE deleted_at IS NULL
  AND (
    username ILIKE '%' || sqlc.arg(search_query) || '%' OR
    email ILIKE '%' || sqlc.arg(search_query) || '%' OR
    phone_number ILIKE '%' || sqlc.arg(search_query) || '%' OR
    first_name ILIKE '%' || sqlc.arg(search_query) || '%' OR
    last_name ILIKE '%' || sqlc.arg(search_query) || '%'
  );
