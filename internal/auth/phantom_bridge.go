package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PhantomBridgePayload struct {
	UserID string `json:"uid"`
	Nonce  string `json:"nonce"`
	Exp    int64  `json:"exp"`
}

func SignPhantomBridgeToken(secret string, userID uuid.UUID, nonce string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("secret required")
	}
	p := PhantomBridgePayload{
		UserID: userID.String(),
		Nonce:  nonce,
		Exp:    time.Now().Add(ttl).Unix(),
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func VerifyPhantomBridgeToken(secret, token string) (*PhantomBridgePayload, error) {
	if secret == "" {
		return nil, errors.New("secret required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expected, sig) {
		return nil, errors.New("invalid token signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var p PhantomBridgePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	if p.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	if _, err := uuid.Parse(p.UserID); err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	if p.Nonce == "" {
		return nil, errors.New("missing nonce")
	}
	return &p, nil
}
