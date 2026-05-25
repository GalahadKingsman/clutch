ALTER TABLE duels DROP CONSTRAINT IF EXISTS duels_human_appeal_fk;
ALTER TABLE duels DROP CONSTRAINT IF EXISTS duels_ai_verdict_fk;

DROP TABLE IF EXISTS human_appeals;
DROP TABLE IF EXISTS ai_verdicts;
DROP TABLE IF EXISTS proofs;

ALTER TABLE duels
    DROP COLUMN IF EXISTS human_appeal_id,
    DROP COLUMN IF EXISTS ai_verdict_id,
    DROP COLUMN IF EXISTS appeal_window_ends_at,
    DROP COLUMN IF EXISTS dispute_opened_by,
    DROP COLUMN IF EXISTS claimed_by;

ALTER TABLE duels DROP CONSTRAINT IF EXISTS duels_status_check;
ALTER TABLE duels ADD CONSTRAINT duels_status_check CHECK (status IN (
    'pending_opponent', 'active', 'awaiting_claim', 'settled', 'cancelled'
));
