package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GalahadKingsman/clutch/internal/auth"
	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) UpsertFromTelegram(ctx context.Context, tg auth.TelegramUser) (*models.User, error) {
	var lastName, username, photo, lang *string
	if tg.LastName != "" {
		lastName = &tg.LastName
	}
	if tg.Username != "" {
		username = &tg.Username
	}
	if tg.PhotoURL != "" {
		photo = &tg.PhotoURL
	}
	if tg.LanguageCode != "" {
		lang = &tg.LanguageCode
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (telegram_id, telegram_username, first_name, last_name, photo_url, language_code, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (telegram_id) DO UPDATE SET
			telegram_username = EXCLUDED.telegram_username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			photo_url = EXCLUDED.photo_url,
			language_code = EXCLUDED.language_code,
			updated_at = NOW()
		RETURNING id, telegram_id, telegram_username, first_name, last_name, photo_url, language_code,
			wallet_address, honor_score, rating, xp, level, wallet_linked_at, created_at, updated_at
	`, tg.ID, username, tg.FirstName, lastName, photo, lang)

	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, telegram_id, telegram_username, first_name, last_name, photo_url, language_code,
			wallet_address, honor_score, rating, xp, level, wallet_linked_at, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) LinkWallet(ctx context.Context, userID uuid.UUID, wallet string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users SET wallet_address = $2, wallet_linked_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, telegram_id, telegram_username, first_name, last_name, photo_url, language_code,
			wallet_address, honor_score, rating, xp, level, wallet_linked_at, created_at, updated_at
	`, userID, wallet)
	return scanUser(row)
}

func (r *UserRepository) CreateWalletNonce(ctx context.Context, userID uuid.UUID, nonce string, expires time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO wallet_link_nonces (user_id, nonce, expires_at) VALUES ($1, $2, $3)
	`, userID, nonce, expires)
	return err
}

func (r *UserRepository) ConsumeWalletNonce(ctx context.Context, userID uuid.UUID, nonce string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE wallet_link_nonces SET used_at = NOW()
		WHERE user_id = $1 AND nonce = $2 AND used_at IS NULL AND expires_at > NOW()
	`, userID, nonce)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.FirstName, &u.LastName, &u.PhotoURL, &u.LanguageCode,
		&u.WalletAddress, &u.HonorScore, &u.Rating, &u.XP, &u.Level, &u.WalletLinkedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
