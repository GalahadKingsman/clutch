package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	DuelDisputed           DuelStatus = "disputed"
	DuelArbitrationUpload  DuelStatus = "arbitration_upload"
	DuelAIJudging          DuelStatus = "ai_judging"
	DuelAppealWindow       DuelStatus = "appeal_window"
	DuelHumanArbitration   DuelStatus = "human_arbitration"
	DuelMutualSettled      DuelStatus = "mutual_settled"
)

type Proof struct {
	ID          uuid.UUID      `json:"id"`
	DuelID      uuid.UUID      `json:"duel_id"`
	UserID      uuid.UUID      `json:"user_id"`
	ProofType   string         `json:"proof_type"`
	StoragePath string         `json:"-"`
	URL         string         `json:"url,omitempty"`
	Caption     *string        `json:"caption,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	ContentHash string         `json:"content_hash,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type AIVerdict struct {
	ID            uuid.UUID   `json:"id"`
	DuelID        uuid.UUID   `json:"duel_id"`
	WinnerID      uuid.UUID   `json:"winner_id"`
	Reasoning     string      `json:"reasoning"`
	Confidence    float64     `json:"confidence"`
	EvidenceRefs  []uuid.UUID `json:"evidence_refs"`
	VerdictHash   string      `json:"verdict_hash"`
	CreatedAt     time.Time   `json:"created_at"`
	AppealEndsAt  *time.Time  `json:"appeal_window_ends_at,omitempty"`
	CanAppeal     bool        `json:"can_appeal"`
	IsWinner      bool        `json:"is_winner,omitempty"`
}

type HumanAppeal struct {
	ID            uuid.UUID  `json:"id"`
	DuelID        uuid.UUID  `json:"duel_id"`
	AppellantID   uuid.UUID  `json:"appellant_id"`
	FeeUSD        float64    `json:"fee_usd"`
	Status        string     `json:"status"`
	SLADeadlineAt *time.Time `json:"sla_deadline_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ClarifyResponse struct {
	NormalizedCondition string `json:"normalized_condition"`
	WinCriterion        string `json:"win_criterion"`
	Tips                string `json:"tips"`
}
