-- name: CreateUserLog :exec
INSERT INTO user_logs (
  user_id,
  action,
  description,
  object_type,
  object_id,
  ip_address,
  user_agent,
  is_success,
  error_message
) VALUES (
  sqlc.arg(user_id),
  sqlc.arg(action),
  sqlc.arg(description),
  sqlc.arg(object_type),
  sqlc.arg(object_id),
  sqlc.arg(ip_address),
  sqlc.arg(user_agent),
  sqlc.arg(is_success),
  sqlc.arg(error_message)
);

-- name: GetUserLogs :many
SELECT * FROM user_logs
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(limitVal) OFFSET sqlc.arg(offsetVal);

-- name: GetUserLogsByAction :many
SELECT * FROM user_logs
WHERE user_id = sqlc.arg(user_id) AND action = sqlc.arg(action)
ORDER BY created_at DESC
LIMIT sqlc.arg(limitVal) OFFSET sqlc.arg(offsetVal);

-- name: GetRecentFailedLogins :many
SELECT * FROM user_logs
WHERE user_id = sqlc.arg(user_id)
  AND action = 'login'
  AND is_success = FALSE
  AND created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- name: CountUserActionLogs :one
SELECT COUNT(*) FROM user_logs
WHERE user_id = sqlc.arg(user_id) AND action = sqlc.arg(action);
