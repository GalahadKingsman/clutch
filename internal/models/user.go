package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	TelegramID       int64      `json:"telegram_id"`
	TelegramUsername *string    `json:"telegram_username,omitempty"`
	FirstName        string     `json:"first_name"`
	LastName         *string    `json:"last_name,omitempty"`
	PhotoURL         *string    `json:"photo_url,omitempty"`
	LanguageCode     *string    `json:"language_code,omitempty"`
	WalletAddress    *string    `json:"wallet_address,omitempty"`
	HonorScore       int        `json:"honor_score"`
	Rating           int        `json:"rating"`
	XP               int        `json:"xp"`
	Level            int        `json:"level"`
	WalletLinkedAt   *time.Time `json:"wallet_linked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (u *User) WalletLinked() bool {
	return u.WalletAddress != nil && *u.WalletAddress != ""
}

type UserResponse struct {
	ID               uuid.UUID `json:"id"`
	TelegramID       int64     `json:"telegram_id"`
	TelegramUsername *string   `json:"telegram_username,omitempty"`
	FirstName        string    `json:"first_name"`
	LastName         *string   `json:"last_name,omitempty"`
	PhotoURL         *string   `json:"photo_url,omitempty"`
	WalletAddress    *string   `json:"wallet_address,omitempty"`
	WalletLinked     bool      `json:"wallet_linked"`
	HonorScore       int       `json:"honor_score"`
	Rating           int       `json:"rating"`
	XP               int       `json:"xp"`
	Level            int       `json:"level"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:               u.ID,
		TelegramID:       u.TelegramID,
		TelegramUsername: u.TelegramUsername,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		PhotoURL:         u.PhotoURL,
		WalletAddress:    u.WalletAddress,
		WalletLinked:     u.WalletLinked(),
		HonorScore:       u.HonorScore,
		Rating:           u.Rating,
		XP:               u.XP,
		Level:            u.Level,
	}
}
