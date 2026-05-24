package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/GalahadKingsman/clutch/internal/auth"
	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/GalahadKingsman/clutch/internal/httputil"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/repository"
)

type AuthHandler struct {
	cfg   *config.Config
	users *repository.UserRepository
}

func NewAuthHandler(cfg *config.Config, users *repository.UserRepository) *AuthHandler {
	return &AuthHandler{cfg: cfg, users: users}
}

type telegramAuthRequest struct {
	InitData string `json:"init_data"`
}

type telegramAuthResponse struct {
	Token *auth.TokenPair    `json:"token"`
	User  any                `json:"user"`
}

func (h *AuthHandler) Telegram(w http.ResponseWriter, r *http.Request) {
	var req telegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.InitData == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "init_data required")
		return
	}

	data, err := auth.ValidateInitData(req.InitData, h.cfg.TelegramBotToken)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid_init_data", err.Error())
		return
	}

	user, err := h.users.UpsertFromTelegram(r.Context(), data.User)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "could not upsert user")
		return
	}

	token, err := auth.IssueAccessToken(h.cfg.JWTSecret, user.ID, user.TelegramID, 7*24*time.Hour)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "could not issue token")
		return
	}

	httputil.JSON(w, http.StatusOK, telegramAuthResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		httputil.Error(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	httputil.JSON(w, http.StatusOK, user.ToResponse())
}

type walletNonceResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

func (h *AuthHandler) WalletNonce(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	nonce, err := randomNonce(16)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	expires := time.Now().Add(10 * time.Minute)
	if err := h.users.CreateWalletNonce(r.Context(), userID, nonce, expires); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	httputil.JSON(w, http.StatusOK, walletNonceResponse{
		Nonce:   nonce,
		Message: auth.BuildSignMessage(nonce),
	})
}

type walletLinkRequest struct {
	WalletAddress string `json:"wallet_address"`
	Signature     string `json:"signature"`
	Nonce         string `json:"nonce"`
}

func (h *AuthHandler) WalletLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	var req walletLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	if req.WalletAddress == "" || req.Signature == "" || req.Nonce == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "wallet_address, signature, nonce required")
		return
	}

	okNonce, err := h.users.ConsumeWalletNonce(r.Context(), userID, req.Nonce)
	if err != nil || !okNonce {
		httputil.Error(w, http.StatusBadRequest, "invalid_nonce", "nonce expired or already used")
		return
	}

	msg := auth.BuildSignMessage(req.Nonce)
	if err := auth.VerifySolanaSignature(req.WalletAddress, msg, req.Signature); err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid_signature", err.Error())
		return
	}

	user, err := h.users.LinkWallet(r.Context(), userID, req.WalletAddress)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	httputil.JSON(w, http.StatusOK, user.ToResponse())
}

func randomNonce(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
