CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               VARCHAR(255) UNIQUE NOT NULL,
    full_name           VARCHAR(255) NOT NULL,
    password_hash       VARCHAR(255) NOT NULL,
    email_verified      BOOLEAN DEFAULT FALSE,
    mfa_type            SMALLINT DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID REFERENCES users(id),
    jti                 UUID UNIQUE NOT NULL,
    expires_at          TIMESTAMPTZ DEFAULT NOW() + INTERVAL '1 day',
    revoked             BOOLEAN DEFAULT FALSE,
    device_fingerprint  VARCHAR(255),
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE mfa_secrets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID REFERENCES users(id),
    totp_secret         VARCHAR(255),
    mfa_type            SMALLINT DEFAULT 0,
    is_verified         BOOLEAN DEFAULT FALSE,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID REFERENCES users(id),
    event_type          VARCHAR(255) NOT NULL,
    ip_address          VARCHAR(255),
    device_fingerprint  VARCHAR(255),
    success             BOOLEAN DEFAULT TRUE,
    metadata            JSONB,
    created_at          TIMESTAMPTZ DEFAULT NOW()      
);
