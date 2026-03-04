CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE secrets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_token    VARCHAR(32) NOT NULL UNIQUE,
    encrypted_data  TEXT NOT NULL,
    iv              VARCHAR(24) NOT NULL,
    salt            VARCHAR(44),
    max_views       INTEGER,
    current_views   INTEGER NOT NULL DEFAULT 0,
    expires_at      TIMESTAMPTZ,
    burn_after_read BOOLEAN NOT NULL DEFAULT false,
    grace_until     TIMESTAMPTZ,
    creator_token   VARCHAR(64) NOT NULL,
    ip_hash         VARCHAR(64),
    ua_hash         VARCHAR(64),
    content_type    VARCHAR(10) NOT NULL DEFAULT 'text',
    is_expired      BOOLEAN NOT NULL DEFAULT false,
    expired_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_secrets_access_token ON secrets(access_token);
CREATE INDEX idx_secrets_expires_at ON secrets(expires_at) WHERE expires_at IS NOT NULL AND is_expired = false;
CREATE INDEX idx_secrets_is_expired ON secrets(is_expired) WHERE is_expired = true;

CREATE TABLE files (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id       UUID NOT NULL UNIQUE REFERENCES secrets(id) ON DELETE CASCADE,
    encrypted_name  TEXT NOT NULL,
    file_size       BIGINT NOT NULL DEFAULT 0,
    original_size   BIGINT NOT NULL,
    storage_key     VARCHAR(512) NOT NULL,
    storage_backend VARCHAR(20) NOT NULL DEFAULT 's3',
    chunk_count     INTEGER NOT NULL DEFAULT 1,
    upload_complete BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_secret_id ON files(secret_id);

CREATE TABLE api_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_prefix      VARCHAR(12) NOT NULL,
    key_hash        VARCHAR(128) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    rate_limit      INTEGER NOT NULL DEFAULT 60,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash) WHERE is_active = true;
