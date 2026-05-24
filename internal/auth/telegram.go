package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TelegramUser from WebApp initData user field.
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
}

// InitData after validation.
type InitData struct {
	User       TelegramUser
	AuthDate   time.Time
	QueryID    string
	RawHash    string
}

const maxInitDataAge = 24 * time.Hour

// ValidateInitData verifies Telegram Mini App initData per official algorithm.
func ValidateInitData(initDataRaw, botToken string) (*InitData, error) {
	if initDataRaw == "" {
		return nil, fmt.Errorf("init data is empty")
	}
	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return nil, fmt.Errorf("parse init data: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("hash missing")
	}

	var pairs []string
	for key, vals := range values {
		if key == "hash" {
			continue
		}
		pairs = append(pairs, key+"="+vals[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	expectedHash := hmacSHA256(secretKey, []byte(dataCheckString))
	if !hmac.Equal([]byte(receivedHash), []byte(hex.EncodeToString(expectedHash))) {
		return nil, fmt.Errorf("invalid init data signature")
	}

	authDateUnix, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date")
	}
	authDate := time.Unix(authDateUnix, 0)
	if time.Since(authDate) > maxInitDataAge {
		return nil, fmt.Errorf("init data expired")
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("user missing in init data")
	}
	var tgUser TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	if tgUser.ID == 0 {
		return nil, fmt.Errorf("invalid telegram user id")
	}

	return &InitData{
		User:     tgUser,
		AuthDate: authDate,
		QueryID:  values.Get("query_id"),
		RawHash:  receivedHash,
	}, nil
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}
