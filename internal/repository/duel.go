package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DuelRepository struct {
	pool *pgxpool.Pool
}

const duelSelectCols = `
	id, on_chain_duel_id, creator_id, opponent_id, condition_text, side_creator, side_opponent,
	stake_usd_each, bank_usd, token_mint, status, deadline_at, winner_id,
	claimed_by, dispute_opened_by, appeal_window_ends_at, ai_verdict_id, human_appeal_id,
	creator_tx, opponent_tx, created_at, updated_at, settled_at
`

func NewDuelRepository(pool *pgxpool.Pool) *DuelRepository {
	return &DuelRepository{pool: pool}
}

type CreateDuelInput struct {
	CreatorID     uuid.UUID
	OpponentID    uuid.UUID
	ConditionText string
	SideCreator   string
	SideOpponent  string
	StakeUSDEach  float64
	DeadlineAt    time.Time
	TokenMint     *string
	CreatorTx     *string
}

func (r *DuelRepository) Create(ctx context.Context, in CreateDuelInput) (*models.Duel, error) {
	bank := in.StakeUSDEach * 2
	row := r.pool.QueryRow(ctx, `
		INSERT INTO duels (
			creator_id, opponent_id, condition_text, side_creator, side_opponent,
			stake_usd_each, bank_usd, token_mint, status, deadline_at, creator_tx
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending_opponent',$9,$10)
		RETURNING ` + duelSelectCols + `
	`, in.CreatorID, in.OpponentID, in.ConditionText, in.SideCreator, in.SideOpponent,
		in.StakeUSDEach, bank, in.TokenMint, in.DeadlineAt, in.CreatorTx)
	return scanDuel(row)
}

func (r *DuelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT ` + duelSelectCols + ` FROM duels WHERE id = $1
	`, id)
	d, err := scanDuel(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

func (r *DuelRepository) Accept(ctx context.Context, duelID, opponentID uuid.UUID, opponentTx *string) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = 'active', opponent_id = $2, opponent_tx = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'pending_opponent' AND opponent_id = $2
		RETURNING ` + duelSelectCols + `
	`, duelID, opponentID, opponentTx)
	return scanDuel(row)
}

func (r *DuelRepository) Cancel(ctx context.Context, duelID, byUser uuid.UUID) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status = 'pending_opponent'
		  AND (creator_id = $2 OR opponent_id = $2)
		RETURNING ` + duelSelectCols + `
	`, duelID, byUser)
	return scanDuel(row)
}

func (r *DuelRepository) SetAwaitingClaim(ctx context.Context, duelID, claimedBy uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE duels SET status = 'awaiting_claim', claimed_by = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'active'
	`, duelID, claimedBy)
	return err
}

func (r *DuelRepository) Settle(ctx context.Context, duelID, winnerID uuid.UUID, status models.DuelStatus) (*models.Duel, error) {
	if status == "" {
		status = models.DuelSettled
	}
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = $3, winner_id = $2, settled_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING `+duelSelectCols, duelID, winnerID, status)
	return scanDuel(row)
}

func (r *DuelRepository) OpenDispute(ctx context.Context, duelID, byUser uuid.UUID) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = 'disputed', dispute_opened_by = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'awaiting_claim'
		RETURNING `+duelSelectCols, duelID, byUser)
	return scanDuel(row)
}

func (r *DuelRepository) SetStatus(ctx context.Context, duelID uuid.UUID, status models.DuelStatus) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING `+duelSelectCols, duelID, status)
	return scanDuel(row)
}

func (r *DuelRepository) SetAppealWindow(ctx context.Context, duelID, verdictID uuid.UUID, endsAt time.Time) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = 'appeal_window', ai_verdict_id = $2, appeal_window_ends_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING `+duelSelectCols, duelID, verdictID, endsAt)
	return scanDuel(row)
}

func (r *DuelRepository) SetHumanAppeal(ctx context.Context, duelID, appealID uuid.UUID) (*models.Duel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE duels SET status = 'human_arbitration', human_appeal_id = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'appeal_window'
		RETURNING `+duelSelectCols, duelID, appealID)
	return scanDuel(row)
}

func (r *DuelRepository) ListForUser(ctx context.Context, userID uuid.UUID, statuses []models.DuelStatus) ([]models.Duel, error) {
	statusStrs := make([]string, len(statuses))
	for i, s := range statuses {
		statusStrs[i] = string(s)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT ` + duelSelectCols + ` FROM duels
		WHERE (creator_id = $1 OR opponent_id = $1)
		  AND status = ANY($2)
		ORDER BY updated_at DESC
	`, userID, statusStrs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Duel
	for rows.Next() {
		d, err := scanDuelRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func (r *DuelRepository) ListIncoming(ctx context.Context, userID uuid.UUID) ([]models.Duel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ` + duelSelectCols + ` FROM duels
		WHERE opponent_id = $1 AND status = 'pending_opponent'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Duel
	for rows.Next() {
		d, err := scanDuelRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func (r *DuelRepository) SetCreatorOnChain(ctx context.Context, duelID uuid.UUID, txSig, pda string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE duels SET creator_tx = $2, on_chain_duel_id = $3, updated_at = NOW()
		WHERE id = $1
	`, duelID, txSig, pda)
	return err
}

func (r *DuelRepository) SetOpponentOnChain(ctx context.Context, duelID uuid.UUID, txSig string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE duels SET opponent_tx = $2, updated_at = NOW()
		WHERE id = $1
	`, duelID, txSig)
	return err
}

func (r *DuelRepository) AddEvent(ctx context.Context, duelID uuid.UUID, eventType string, payload []byte) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO duel_events (duel_id, event_type, payload) VALUES ($1, $2, $3)
	`, duelID, eventType, payload)
	return err
}

func scanDuel(row pgx.Row) (*models.Duel, error) {
	var d models.Duel
	var status string
	err := row.Scan(
		&d.ID, &d.OnChainDuelID, &d.CreatorID, &d.OpponentID, &d.ConditionText, &d.SideCreator, &d.SideOpponent,
		&d.StakeUSDEach, &d.BankUSD, &d.TokenMint, &status, &d.DeadlineAt, &d.WinnerID,
		&d.ClaimedBy, &d.DisputeOpenedBy, &d.AppealWindowEndsAt, &d.AIVerdictID, &d.HumanAppealID,
		&d.CreatorTx, &d.OpponentTx, &d.CreatedAt, &d.UpdatedAt, &d.SettledAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan duel: %w", err)
	}
	d.Status = models.DuelStatus(status)
	return &d, nil
}

func scanDuelRows(rows pgx.Rows) (*models.Duel, error) {
	var d models.Duel
	var status string
	err := rows.Scan(
		&d.ID, &d.OnChainDuelID, &d.CreatorID, &d.OpponentID, &d.ConditionText, &d.SideCreator, &d.SideOpponent,
		&d.StakeUSDEach, &d.BankUSD, &d.TokenMint, &status, &d.DeadlineAt, &d.WinnerID,
		&d.ClaimedBy, &d.DisputeOpenedBy, &d.AppealWindowEndsAt, &d.AIVerdictID, &d.HumanAppealID,
		&d.CreatorTx, &d.OpponentTx, &d.CreatedAt, &d.UpdatedAt, &d.SettledAt,
	)
	if err != nil {
		return nil, err
	}
	d.Status = models.DuelStatus(status)
	return &d, nil
}
