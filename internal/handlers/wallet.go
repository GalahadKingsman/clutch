package handlers

import (
	"net/http"

	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/GalahadKingsman/clutch/internal/httputil"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/GalahadKingsman/clutch/internal/solana"
	solanago "github.com/gagliardetto/solana-go"
)

type WalletHandler struct {
	cfg   *config.Config
	users *repository.UserRepository
	chain *solana.Client
}

func NewWalletHandler(cfg *config.Config, users *repository.UserRepository, chain *solana.Client) *WalletHandler {
	return &WalletHandler{cfg: cfg, users: users, chain: chain}
}

func (h *WalletHandler) Balances(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	if h.chain == nil {
		httputil.Error(w, http.StatusServiceUnavailable, "chain_unavailable", "")
		return
	}

	u, err := h.users.GetByID(r.Context(), userID)
	if err != nil || u == nil || u.WalletAddress == nil {
		httputil.Error(w, http.StatusBadRequest, "no_wallet", "link wallet first")
		return
	}

	owner, err := solanago.PublicKeyFromBase58(*u.WalletAddress)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_wallet", "")
		return
	}

	sol, err := h.chain.SOLBalance(r.Context(), owner)
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "rpc_error", err.Error())
		return
	}

	mint := h.cfg.UsdcMintDevnet
	usdc, _ := h.chain.USDCBalance(r.Context(), owner, mint)

	httputil.JSON(w, http.StatusOK, map[string]any{
		"wallet": *u.WalletAddress,
		"sol":    sol,
		"usdc":   usdc,
		"mint":   mint,
		"network": "devnet",
	})
}
