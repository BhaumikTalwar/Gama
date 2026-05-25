-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  user_id,
  token_hash,
  expires_at,
  user_agent,
  ip_address,
  device_id
) VALUES (
  sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at), sqlc.arg(user_agent), sqlc.arg(ip_address), sqlc.arg(device_id)
) RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked = FALSE
  AND expires_at > NOW();

-- name: GetRefreshTokenByID :one
SELECT * FROM refresh_tokens
WHERE id = sqlc.arg(id);

-- name: UpdateRefreshTokenLastUsed :exec
UPDATE refresh_tokens
SET last_used_at = NOW()
WHERE id = sqlc.arg(id);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
  revoked = TRUE,
  revoked_at = NOW(),
  revoked_reason = sqlc.arg(revoked_reason)
WHERE id = sqlc.arg(id);

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET
  revoked = TRUE,
  revoked_at = NOW(),
  revoked_reason = sqlc.arg(revoked_reason)
WHERE token_hash = sqlc.arg(token_hash);

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET
  revoked = TRUE,
  revoked_at = NOW(),
  revoked_reason = 'user_logout_all'
WHERE user_id = sqlc.arg(user_id) AND revoked = FALSE;

-- name: RevokeUserTokensByDevice :exec
UPDATE refresh_tokens
SET
  revoked = TRUE,
  revoked_at = NOW(),
  revoked_reason = 'device_logout'
WHERE user_id = sqlc.arg(user_id)
  AND device_id = sqlc.arg(device_id)
  AND revoked = FALSE;

-- name: DeleteExpiredTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < NOW();

-- name: GetUserActiveTokens :many
SELECT * FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id)
  AND revoked = FALSE
  AND expires_at > NOW()
ORDER BY last_used_at DESC;

-- name: GetSuspiciousTokenActivity :many
SELECT 
  user_id,
  COUNT(DISTINCT ip_address) as unique_ips,
  COUNT(*) as token_count,
  MAX(created_at) as latest_token
FROM refresh_tokens
WHERE created_at > NOW() - interval '1 hour'
  AND revoked = FALSE
GROUP BY user_id
HAVING COUNT(*) > sqlc.arg(token_threshold) 
  OR COUNT(DISTINCT ip_address) > sqlc.arg(ip_threshold)
ORDER BY token_count DESC;

-- name: GetUserTokenStats :one
SELECT 
  COUNT(*) as total_tokens,
  COUNT(*) FILTER (WHERE revoked = FALSE AND expires_at > NOW()) as active_tokens,
  COUNT(*) FILTER (WHERE revoked = TRUE) as revoked_tokens,
  COUNT(*) FILTER (WHERE expires_at <= NOW()) as expired_tokens,
  MAX(last_used_at) as last_token_use,
  COUNT(DISTINCT device_id) FILTER (WHERE device_id IS NOT NULL) as unique_devices
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id);
