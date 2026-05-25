package repository

import (
	"context"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) Insert(ctx context.Context, duelID uuid.UUID, userID *uuid.UUID, body string, isSystem bool) (*models.ChatMessage, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO chat_messages (duel_id, user_id, body, is_system)
		VALUES ($1, $2, $3, $4)
		RETURNING id, duel_id, user_id, body, is_system, created_at
	`, duelID, userID, body, isSystem)
	var m models.ChatMessage
	err := row.Scan(&m.ID, &m.DuelID, &m.UserID, &m.Body, &m.IsSystem, &m.CreatedAt)
	return &m, err
}

func (r *ChatRepository) List(ctx context.Context, duelID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, duel_id, user_id, body, is_system, created_at
		FROM chat_messages WHERE duel_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, duelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ChatMessage
	for rows.Next() {
		var m models.ChatMessage
		if err := rows.Scan(&m.ID, &m.DuelID, &m.UserID, &m.Body, &m.IsSystem, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
