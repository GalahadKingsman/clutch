package repository

import (
	"context"
	"encoding/json"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProofRepository struct {
	pool *pgxpool.Pool
}

func NewProofRepository(pool *pgxpool.Pool) *ProofRepository {
	return &ProofRepository{pool: pool}
}

type InsertProofInput struct {
	DuelID      uuid.UUID
	UserID      uuid.UUID
	ProofType   string
	StoragePath string
	Caption     *string
	Metadata    map[string]any
	ContentHash string
}

func (r *ProofRepository) Insert(ctx context.Context, in InsertProofInput) (*models.Proof, error) {
	meta, _ := json.Marshal(in.Metadata)
	row := r.pool.QueryRow(ctx, `
		INSERT INTO proofs (duel_id, user_id, proof_type, storage_path, caption, metadata, content_hash)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, duel_id, user_id, proof_type, storage_path, caption, metadata, content_hash, created_at
	`, in.DuelID, in.UserID, in.ProofType, in.StoragePath, in.Caption, meta, in.ContentHash)
	return scanProof(row)
}

func (r *ProofRepository) ListByDuel(ctx context.Context, duelID uuid.UUID) ([]models.Proof, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, duel_id, user_id, proof_type, storage_path, caption, metadata, content_hash, created_at
		FROM proofs WHERE duel_id = $1 ORDER BY created_at ASC
	`, duelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Proof
	for rows.Next() {
		p, err := scanProofRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *ProofRepository) CountByUser(ctx context.Context, duelID, userID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM proofs WHERE duel_id = $1 AND user_id = $2
	`, duelID, userID).Scan(&n)
	return n, err
}

func (r *ProofRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Proof, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, duel_id, user_id, proof_type, storage_path, caption, metadata, content_hash, created_at
		FROM proofs WHERE id = $1
	`, id)
	return scanProof(row)
}

func scanProof(row pgx.Row) (*models.Proof, error) {
	var p models.Proof
	var meta []byte
	var caption *string
	err := row.Scan(
		&p.ID, &p.DuelID, &p.UserID, &p.ProofType, &p.StoragePath, &caption, &meta, &p.ContentHash, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Caption = caption
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &p.Metadata)
	}
	return &p, nil
}

func scanProofRows(rows pgx.Rows) (*models.Proof, error) {
	var p models.Proof
	var meta []byte
	var caption *string
	err := rows.Scan(
		&p.ID, &p.DuelID, &p.UserID, &p.ProofType, &p.StoragePath, &caption, &meta, &p.ContentHash, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Caption = caption
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &p.Metadata)
	}
	return &p, nil
}
