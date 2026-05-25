-- name: GetUsersByIDs :many
SELECT * FROM users
WHERE id = ANY(sqlc.arg('user_ids')::UUID[]) AND deleted_at IS NULL;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone_number = sqlc.arg(phone_number) AND deleted_at IS NULL;

-- name: VerifyUserPhone :exec
UPDATE users
SET phone_verified = TRUE
WHERE id = sqlc.arg(id);

-- name: UpdateUserEmail :exec
UPDATE users
SET email = sqlc.arg(email), email_verified = FALSE
WHERE id = sqlc.arg(id);

-- name: UpdateUserPhone :exec
UPDATE users
SET phone_number = sqlc.arg(phone_number), phone_verified = FALSE
WHERE id = sqlc.arg(id);