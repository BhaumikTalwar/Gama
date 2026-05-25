-- name: CreateVerificationToken :one
INSERT INTO verification_tokens (
    user_id,
    token_hash,
    type,
    expires_at,
    external_id
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(token_type),
    sqlc.arg(expires_at),
    sqlc.arg(external_id)
)
RETURNING *;

-- name: GetVerificationTokenByHash :one
SELECT * FROM verification_tokens
WHERE 
    token_hash = sqlc.arg(token_hash)
    AND type = sqlc.arg(token_type)
    AND used_at IS NULL
    AND expires_at > NOW()
LIMIT 1;

-- name: GetLatestVerificationTokenForUser :one
SELECT * FROM verification_tokens
WHERE
    user_id = sqlc.arg(user_id)
    AND type = sqlc.arg(token_type)
    AND used_at IS NULL
    AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: GetVerificationTokenByExternalID :one
SELECT * FROM verification_tokens
WHERE
    user_id = sqlc.arg(user_id)
    AND type = sqlc.arg(token_type)
    AND external_id = sqlc.arg(external_id)
    AND used_at IS NULL
    AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkTokenUsed :exec
UPDATE verification_tokens
SET used_at = NOW()
WHERE id = sqlc.arg(id);

-- name: CleanupExpiredTokens :exec
DELETE FROM verification_tokens
WHERE expires_at < NOW() 
AND (used_at IS NOT NULL OR expires_at < NOW() - INTERVAL '1 day');
