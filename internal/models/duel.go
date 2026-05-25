package models

import (
	"time"

	"github.com/google/uuid"
)

type DuelStatus string

const (
	DuelPendingOpponent DuelStatus = "pending_opponent"
	DuelActive          DuelStatus = "active"
	DuelAwaitingClaim   DuelStatus = "awaiting_claim"
	DuelSettled         DuelStatus = "settled"
	DuelCancelled       DuelStatus = "cancelled"
)

type Duel struct {
	ID             uuid.UUID  `json:"id"`
	OnChainDuelID  *string    `json:"on_chain_duel_id,omitempty"`
	CreatorID      uuid.UUID  `json:"creator_id"`
	OpponentID     *uuid.UUID `json:"opponent_id,omitempty"`
	ConditionText  string     `json:"condition_text"`
	SideCreator    string     `json:"side_creator"`
	SideOpponent   string     `json:"side_opponent"`
	StakeUSDEach   float64    `json:"stake_usd_each"`
	BankUSD        float64    `json:"bank_usd"`
	TokenMint      *string    `json:"token_mint,omitempty"`
	Status         DuelStatus `json:"status"`
	DeadlineAt     time.Time  `json:"deadline_at"`
	WinnerID            *uuid.UUID `json:"winner_id,omitempty"`
	ClaimedBy           *uuid.UUID `json:"claimed_by,omitempty"`
	DisputeOpenedBy     *uuid.UUID `json:"dispute_opened_by,omitempty"`
	AppealWindowEndsAt  *time.Time `json:"appeal_window_ends_at,omitempty"`
	AIVerdictID         *uuid.UUID `json:"ai_verdict_id,omitempty"`
	HumanAppealID       *uuid.UUID `json:"human_appeal_id,omitempty"`
	CreatorTx           *string    `json:"creator_tx,omitempty"`
	OpponentTx          *string    `json:"opponent_tx,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	SettledAt      *time.Time `json:"settled_at,omitempty"`
}

type DuelParticipant struct {
	ID               uuid.UUID `json:"id"`
	FirstName        string    `json:"first_name"`
	TelegramUsername *string   `json:"telegram_username,omitempty"`
	PhotoURL         *string   `json:"photo_url,omitempty"`
}

type DuelCard struct {
	Duel
	Creator  DuelParticipant  `json:"creator"`
	Opponent *DuelParticipant `json:"opponent,omitempty"`
}

type ChatMessage struct {
	ID        uuid.UUID  `json:"id"`
	DuelID    uuid.UUID  `json:"duel_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Body      string     `json:"body"`
	IsSystem  bool       `json:"is_system"`
	CreatedAt time.Time  `json:"created_at"`
}
