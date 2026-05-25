package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerdictRepository struct {
	pool *pgxpool.Pool
}

func NewVerdictRepository(pool *pgxpool.Pool) *VerdictRepository {
	return &VerdictRepository{pool: pool}
}

type InsertVerdictInput struct {
	DuelID       uuid.UUID
	WinnerID     uuid.UUID
	Reasoning    string
	Confidence   float64
	EvidenceRefs []uuid.UUID
	VerdictHash  string
}

func (r *VerdictRepository) Insert(ctx context.Context, in InsertVerdictInput) (*models.AIVerdict, error) {
	refsStrs := make([]string, len(in.EvidenceRefs))
	for i, id := range in.EvidenceRefs {
		refsStrs[i] = id.String()
	}
	refs, _ := json.Marshal(refsStrs)
	row := r.pool.QueryRow(ctx, `
		INSERT INTO ai_verdicts (duel_id, winner_id, reasoning, confidence, evidence_refs, verdict_hash)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, duel_id, winner_id, reasoning, confidence, evidence_refs, verdict_hash, created_at
	`, in.DuelID, in.WinnerID, in.Reasoning, in.Confidence, refs, in.VerdictHash)
	return scanVerdict(row)
}

func (r *VerdictRepository) GetByDuel(ctx context.Context, duelID uuid.UUID) (*models.AIVerdict, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, duel_id, winner_id, reasoning, confidence, evidence_refs, verdict_hash, created_at
		FROM ai_verdicts WHERE duel_id = $1 ORDER BY created_at DESC LIMIT 1
	`, duelID)
	v, err := scanVerdict(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return v, err
}

func (r *VerdictRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AIVerdict, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, duel_id, winner_id, reasoning, confidence, evidence_refs, verdict_hash, created_at
		FROM ai_verdicts WHERE id = $1
	`, id)
	return scanVerdict(row)
}

type InsertAppealInput struct {
	DuelID        uuid.UUID
	AppellantID   uuid.UUID
	FeeUSD        float64
	SLADeadlineAt interface{}
}

func (r *VerdictRepository) InsertAppeal(ctx context.Context, duelID, appellantID uuid.UUID, feeUSD float64, slaDeadline interface{}) (*models.HumanAppeal, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO human_appeals (duel_id, appellant_id, fee_usd, sla_deadline_at)
		VALUES ($1,$2,$3,$4)
		RETURNING id, duel_id, appellant_id, fee_usd, status, sla_deadline_at, created_at
	`, duelID, appellantID, feeUSD, slaDeadline)
	var a models.HumanAppeal
	err := row.Scan(&a.ID, &a.DuelID, &a.AppellantID, &a.FeeUSD, &a.Status, &a.SLADeadlineAt, &a.CreatedAt)
	return &a, err
}

func scanVerdict(row pgx.Row) (*models.AIVerdict, error) {
	var v models.AIVerdict
	var refs []byte
	err := row.Scan(&v.ID, &v.DuelID, &v.WinnerID, &v.Reasoning, &v.Confidence, &refs, &v.VerdictHash, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	if len(refs) > 0 {
		var ids []string
		_ = json.Unmarshal(refs, &ids)
		for _, id := range ids {
			if u, err := uuid.Parse(id); err == nil {
				v.EvidenceRefs = append(v.EvidenceRefs, u)
			}
		}
	}
	return &v, nil
}
