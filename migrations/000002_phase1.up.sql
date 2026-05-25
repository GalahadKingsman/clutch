CREATE TABLE invite_codes (
    code        TEXT PRIMARY KEY,
    inviter_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX invite_codes_inviter_idx ON invite_codes (inviter_id);

CREATE TABLE duels (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    on_chain_duel_id    TEXT,
    creator_id          UUID NOT NULL REFERENCES users(id),
    opponent_id         UUID REFERENCES users(id),
    condition_text      TEXT NOT NULL,
    side_creator        TEXT NOT NULL,
    side_opponent       TEXT NOT NULL,
    stake_usd_each      NUMERIC(18, 2) NOT NULL,
    bank_usd            NUMERIC(18, 2) NOT NULL,
    token_mint          TEXT,
    status              TEXT NOT NULL DEFAULT 'pending_opponent'
        CHECK (status IN (
            'pending_opponent', 'active', 'awaiting_claim',
            'settled', 'cancelled'
        )),
    deadline_at         TIMESTAMPTZ NOT NULL,
    winner_id           UUID REFERENCES users(id),
    creator_tx          TEXT,
    opponent_tx         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at          TIMESTAMPTZ
);

CREATE INDEX duels_status_deadline_idx ON duels (status, deadline_at);
CREATE INDEX duels_creator_idx ON duels (creator_id);
CREATE INDEX duels_opponent_idx ON duels (opponent_id);

CREATE TABLE chat_messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duel_id     UUID NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id),
    body        TEXT NOT NULL,
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX chat_messages_duel_idx ON chat_messages (duel_id, created_at);

CREATE TABLE duel_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duel_id     UUID NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX duel_events_duel_idx ON duel_events (duel_id, created_at DESC);

CREATE TABLE activity_feed (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type  TEXT NOT NULL,
    ref_id      UUID,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX activity_feed_user_created_idx ON activity_feed (user_id, created_at DESC);
