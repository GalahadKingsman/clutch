-- Phase 2: dispute, proofs, AI verdict, human appeal

ALTER TABLE duels DROP CONSTRAINT IF EXISTS duels_status_check;

ALTER TABLE duels ADD CONSTRAINT duels_status_check CHECK (status IN (
    'pending_opponent', 'active', 'awaiting_claim', 'disputed',
    'arbitration_upload', 'ai_judging', 'appeal_window', 'human_arbitration',
    'settled', 'cancelled', 'mutual_settled'
));

ALTER TABLE duels
    ADD COLUMN IF NOT EXISTS claimed_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS dispute_opened_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS appeal_window_ends_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ai_verdict_id UUID,
    ADD COLUMN IF NOT EXISTS human_appeal_id UUID;

CREATE TABLE proofs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duel_id         UUID NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),
    proof_type      TEXT NOT NULL DEFAULT 'image'
        CHECK (proof_type IN ('image', 'video', 'text', 'geo')),
    storage_path    TEXT NOT NULL,
    caption         TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    content_hash    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX proofs_duel_idx ON proofs (duel_id, created_at);
CREATE INDEX proofs_user_duel_idx ON proofs (duel_id, user_id);

CREATE TABLE ai_verdicts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duel_id         UUID NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    winner_id       UUID NOT NULL REFERENCES users(id),
    reasoning       TEXT NOT NULL,
    confidence      NUMERIC(5, 4) NOT NULL,
    evidence_refs   JSONB NOT NULL DEFAULT '[]',
    verdict_hash    TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ai_verdicts_duel_idx ON ai_verdicts (duel_id);

CREATE TABLE human_appeals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duel_id         UUID NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    appellant_id    UUID NOT NULL REFERENCES users(id),
    fee_usd         NUMERIC(18, 2) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed_ai', 'overturned', 'expired')),
    decision_note   TEXT,
    sla_deadline_at TIMESTAMPTZ,
    decided_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX human_appeals_duel_idx ON human_appeals (duel_id);
CREATE INDEX human_appeals_status_idx ON human_appeals (status);

ALTER TABLE duels
    ADD CONSTRAINT duels_ai_verdict_fk
        FOREIGN KEY (ai_verdict_id) REFERENCES ai_verdicts(id),
    ADD CONSTRAINT duels_human_appeal_fk
        FOREIGN KEY (human_appeal_id) REFERENCES human_appeals(id);
