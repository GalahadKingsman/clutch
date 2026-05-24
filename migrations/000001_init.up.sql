CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id     BIGINT NOT NULL UNIQUE,
    telegram_username TEXT,
    first_name      TEXT NOT NULL DEFAULT '',
    last_name       TEXT,
    photo_url       TEXT,
    language_code   TEXT,
    wallet_address  TEXT UNIQUE,
    honor_score     INT NOT NULL DEFAULT 100,
    rating          INT NOT NULL DEFAULT 1000,
    xp              INT NOT NULL DEFAULT 0,
    level           INT NOT NULL DEFAULT 1,
    wallet_linked_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_telegram_username_idx
    ON users (LOWER(telegram_username))
    WHERE telegram_username IS NOT NULL AND telegram_username <> '';

CREATE UNIQUE INDEX users_wallet_address_idx
    ON users (wallet_address)
    WHERE wallet_address IS NOT NULL;

CREATE INDEX users_name_trgm_idx
    ON users USING gin ((first_name || ' ' || COALESCE(last_name, '')) gin_trgm_ops);

CREATE TABLE friendships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'accepted', 'blocked')),
    contact_alias   TEXT,
    invite_code     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, friend_id)
);

CREATE INDEX friendships_user_status_idx ON friendships (user_id, status);

CREATE TABLE wallet_link_nonces (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nonce           TEXT NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX wallet_link_nonces_user_idx ON wallet_link_nonces (user_id);
