package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

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

type phantomBridgePrepareResponse struct {
	PhantomURL string `json:"phantom_url"`
	PageURL    string `json:"page_url"`
}

// PhantomBridgePrepare — одноразовая ссылка для привязки в in-app браузере Phantom (Telegram).
func (h *AuthHandler) PhantomBridgePrepare(w http.ResponseWriter, r *http.Request) {
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
	token, err := auth.SignPhantomBridgeToken(h.cfg.JWTSecret, userID, nonce, 10*time.Minute)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	appURL := strings.TrimRight(h.cfg.MiniAppPublicURL, "/")
	if appURL == "" || appURL == "https://app.example.com" {
		appURL = strings.TrimRight(h.cfg.CORSOrigins[0], "/")
	}
	pageURL := fmt.Sprintf("%s/tg-wallet?token=%s", appURL, url.QueryEscape(token))
	phantomURL := fmt.Sprintf(
		"https://phantom.app/ul/browse/%s?ref=%s",
		url.QueryEscape(pageURL),
		url.QueryEscape(appURL),
	)
	httputil.JSON(w, http.StatusOK, phantomBridgePrepareResponse{
		PhantomURL: phantomURL,
		PageURL:    pageURL,
	})
}

func (h *AuthHandler) PhantomBridgeSession(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "token required")
		return
	}
	payload, err := auth.VerifyPhantomBridgeToken(h.cfg.JWTSecret, token)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid_token", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, walletNonceResponse{
		Nonce:   payload.Nonce,
		Message: auth.BuildSignMessage(payload.Nonce),
	})
}

type phantomBridgeLinkRequest struct {
	Token         string `json:"token"`
	WalletAddress string `json:"wallet_address"`
	Signature     string `json:"signature"`
}

func (h *AuthHandler) PhantomBridgeLink(w http.ResponseWriter, r *http.Request) {
	var req phantomBridgeLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	if req.Token == "" || req.WalletAddress == "" || req.Signature == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "token, wallet_address, signature required")
		return
	}
	payload, err := auth.VerifyPhantomBridgeToken(h.cfg.JWTSecret, req.Token)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid_token", err.Error())
		return
	}
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	okNonce, err := h.users.ConsumeWalletNonce(r.Context(), userID, payload.Nonce)
	if err != nil || !okNonce {
		httputil.Error(w, http.StatusBadRequest, "invalid_nonce", "nonce expired or already used")
		return
	}
	msg := auth.BuildSignMessage(payload.Nonce)
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
