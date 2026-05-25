-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_used_at TIMESTAMPTZ,
  revoked BOOLEAN NOT NULL DEFAULT FALSE,
  revoked_at TIMESTAMPTZ,
  revoked_reason TEXT,
  user_agent TEXT,
  ip_address INET,
  device_id TEXT,
  CONSTRAINT valid_revoked_at CHECK (
    (revoked = FALSE AND revoked_at IS NULL) OR 
    (revoked = TRUE AND revoked_at IS NOT NULL)
  )
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash) WHERE revoked = FALSE;
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_device_id ON refresh_tokens(device_id) WHERE device_id IS NOT NULL;

CREATE TYPE mfa_type AS ENUM ('totp', 'sms');
CREATE TYPE token_type AS ENUM (
    'email_verification', 
    'password_reset', 
    'phone_verification', 
    'email_reverification',
    'sms_otp'
);

CREATE TABLE IF NOT EXISTS user_mfa_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret_key TEXT, 
    method mfa_type NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    backup_codes TEXT[], 
    phone_number VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    type token_type NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    external_id TEXT,
    used_at TIMESTAMPTZ,
    
    UNIQUE (user_id, type, token_hash)
);

CREATE INDEX idx_verification_tokens_user_id ON verification_tokens(user_id);
CREATE INDEX idx_verification_tokens_token_hash ON verification_tokens(token_hash) WHERE used_at IS NULL;
CREATE INDEX idx_verification_tokens_expires_at ON verification_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_lookup ON verification_tokens(token_hash, type) WHERE used_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_verification_tokens_cleanup ON verification_tokens(expires_at) WHERE used_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS verification_tokens;
DROP TABLE IF EXISTS user_mfa_settings;
DROP TABLE IF EXISTS refresh_tokens;
