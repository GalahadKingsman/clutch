package repository

import (
	"context"
	"fmt"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FriendRepository struct {
	pool *pgxpool.Pool
}

func NewFriendRepository(pool *pgxpool.Pool) *FriendRepository {
	return &FriendRepository{pool: pool}
}

func (r *FriendRepository) AreFriends(ctx context.Context, a, b uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM friendships
			WHERE status = 'accepted'
			  AND ((user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1))
		)
	`, a, b).Scan(&ok)
	return ok, err
}

func (r *FriendRepository) ListFriends(ctx context.Context, userID uuid.UUID) ([]models.FriendCard, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.telegram_id, u.telegram_username, u.first_name, u.last_name, u.photo_url,
			u.wallet_address, u.honor_score, u.rating, u.xp, u.level,
			f.contact_alias, f.created_at
		FROM friendships f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = $1 AND f.status = 'accepted'
		ORDER BY u.first_name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.FriendCard
	for rows.Next() {
		var u models.User
		var card models.FriendCard
		if err := rows.Scan(
			&u.ID, &u.TelegramID, &u.TelegramUsername, &u.FirstName, &u.LastName, &u.PhotoURL,
			&u.WalletAddress, &u.HonorScore, &u.Rating, &u.XP, &u.Level,
			&card.ContactAlias, &card.Since,
		); err != nil {
			return nil, err
		}
		card.User = u.ToResponse()
		out = append(out, card)
	}
	return out, rows.Err()
}

func (r *FriendRepository) SearchFriends(ctx context.Context, userID uuid.UUID, q string) ([]models.FriendCard, error) {
	pattern := "%" + q + "%"
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.telegram_id, u.telegram_username, u.first_name, u.last_name, u.photo_url,
			u.wallet_address, u.honor_score, u.rating, u.xp, u.level,
			f.contact_alias, f.created_at
		FROM friendships f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = $1 AND f.status = 'accepted'
		  AND (
			f.contact_alias ILIKE $2
			OR u.first_name ILIKE $2
			OR u.last_name ILIKE $2
			OR u.telegram_username ILIKE $2
		  )
		ORDER BY u.first_name
		LIMIT 20
	`, userID, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.FriendCard
	for rows.Next() {
		var u models.User
		var card models.FriendCard
		if err := rows.Scan(
			&u.ID, &u.TelegramID, &u.TelegramUsername, &u.FirstName, &u.LastName, &u.PhotoURL,
			&u.WalletAddress, &u.HonorScore, &u.Rating, &u.XP, &u.Level,
			&card.ContactAlias, &card.Since,
		); err != nil {
			return nil, err
		}
		card.User = u.ToResponse()
		out = append(out, card)
	}
	return out, rows.Err()
}

func (r *FriendRepository) SearchUsers(ctx context.Context, q string, excludeID uuid.UUID) ([]models.User, error) {
	pattern := "%" + q + "%"
	rows, err := r.pool.Query(ctx, `
		SELECT id, telegram_id, telegram_username, first_name, last_name, photo_url, language_code,
			wallet_address, honor_score, rating, xp, level, wallet_linked_at, created_at, updated_at
		FROM users
		WHERE id != $1
		  AND (
			telegram_username ILIKE $2
			OR first_name ILIKE $2
			OR last_name ILIKE $2
		  )
		ORDER BY first_name
		LIMIT 20
	`, excludeID, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		u, err := scanUserRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (r *FriendRepository) CreateInviteCode(ctx context.Context, code string, inviterID uuid.UUID, expiresAt interface{}) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invite_codes (code, inviter_id, expires_at) VALUES ($1, $2, $3)
	`, code, inviterID, expiresAt)
	return err
}

func (r *FriendRepository) AcceptInvite(ctx context.Context, code string, joinerID uuid.UUID) (*uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var inviterID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT inviter_id FROM invite_codes WHERE code = $1 AND expires_at > NOW()
	`, code).Scan(&inviterID)
	if err != nil {
		return nil, fmt.Errorf("invalid invite")
	}
	if inviterID == joinerID {
		return nil, fmt.Errorf("cannot invite yourself")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'accepted')
		ON CONFLICT (user_id, friend_id) DO UPDATE SET status = 'accepted'
	`, inviterID, joinerID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'accepted')
		ON CONFLICT (user_id, friend_id) DO UPDATE SET status = 'accepted'
	`, joinerID, inviterID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &inviterID, nil
}

func scanUserRows(rows pgx.Rows) (*models.User, error) {
	var u models.User
	return &u, rows.Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.FirstName, &u.LastName, &u.PhotoURL, &u.LanguageCode,
		&u.WalletAddress, &u.HonorScore, &u.Rating, &u.XP, &u.Level, &u.WalletLinkedAt, &u.CreatedAt, &u.UpdatedAt,
	)
}
