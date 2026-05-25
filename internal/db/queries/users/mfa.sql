-- name: GetUserMFASettings :one
SELECT 
    sqlc.embed(settings),
    u.email,
    u.first_name
FROM user_mfa_settings settings
JOIN users u ON u.id = settings.user_id
WHERE settings.user_id = sqlc.arg(user_id);

-- name: UpsertMFASettings :one
INSERT INTO user_mfa_settings (
    user_id,
    secret_key,
    method,
    phone_number,
    enabled,
    backup_codes
) VALUES (
    sqlc.arg(user_id),
    sqlc.narg(secret_key),   
    sqlc.arg(method),
    sqlc.narg(phone_number), 
    sqlc.arg(enabled),       
    sqlc.narg(backup_codes)  
)
ON CONFLICT (user_id) DO UPDATE
SET 
    secret_key = COALESCE(EXCLUDED.secret_key, user_mfa_settings.secret_key),
    method = EXCLUDED.method,
    phone_number = COALESCE(EXCLUDED.phone_number, user_mfa_settings.phone_number),
    updated_at = NOW()
RETURNING *;

-- name: EnableMFA :exec
UPDATE user_mfa_settings
SET 
    enabled = TRUE,
    backup_codes = sqlc.narg(backup_codes), 
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id);

-- name: UpdateMFAMethod :exec
UPDATE user_mfa_settings
SET 
    method = sqlc.arg(method),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id);

-- name: DisableMFA :exec
DELETE FROM user_mfa_settings
WHERE user_id = sqlc.arg(user_id);
