package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GalahadKingsman/clutch/internal/auth"
	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/GalahadKingsman/clutch/internal/httputil"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/GalahadKingsman/clutch/internal/service"
	"github.com/GalahadKingsman/clutch/internal/solana"
	"github.com/GalahadKingsman/clutch/internal/telegram"
	"github.com/GalahadKingsman/clutch/internal/ws"
	solanago "github.com/gagliardetto/solana-go"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var duelUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type DuelHandler struct {
	cfg      *config.Config
	duels    *repository.DuelRepository
	friends  *repository.FriendRepository
	users    *repository.UserRepository
	chat     *repository.ChatRepository
	cards    *service.DuelCardService
	notify   *telegram.Notifier
	hub      *ws.Hub
	chain    *solana.Client
}

func NewDuelHandler(
	cfg *config.Config,
	duels *repository.DuelRepository,
	friends *repository.FriendRepository,
	users *repository.UserRepository,
	chat *repository.ChatRepository,
	cards *service.DuelCardService,
	notify *telegram.Notifier,
	hub *ws.Hub,
	chain *solana.Client,
) *DuelHandler {
	return &DuelHandler{cfg: cfg, duels: duels, friends: friends, users: users, chat: chat, cards: cards, notify: notify, hub: hub, chain: chain}
}

type createDuelRequest struct {
	OpponentID    string  `json:"opponent_id"`
	ConditionText string  `json:"condition_text"`
	SideCreator   string  `json:"side_creator"`
	SideOpponent  string  `json:"side_opponent"`
	StakeUSDEach  float64 `json:"stake_usd_each"`
	DeadlineHours int     `json:"deadline_hours"`
	CreatorTx     *string `json:"creator_tx"`
}

func (h *DuelHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	var req createDuelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	oppID, err := uuid.Parse(req.OpponentID)
	if err != nil || req.ConditionText == "" || req.StakeUSDEach <= 0 {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "invalid fields")
		return
	}
	friends, err := h.friends.AreFriends(r.Context(), userID, oppID)
	if err != nil || !friends {
		httputil.Error(w, http.StatusForbidden, "not_friends", "only friends can duel")
		return
	}
	hours := req.DeadlineHours
	if hours <= 0 {
		hours = 24
	}
	d, err := h.duels.Create(r.Context(), repository.CreateDuelInput{
		CreatorID:     userID,
		OpponentID:    oppID,
		ConditionText: req.ConditionText,
		SideCreator:   req.SideCreator,
		SideOpponent:  req.SideOpponent,
		StakeUSDEach:  req.StakeUSDEach,
		DeadlineAt:    time.Now().Add(time.Duration(hours) * time.Hour),
		CreatorTx:     req.CreatorTx,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	_, _ = h.chat.Insert(r.Context(), d.ID, nil, "⚖️ Судья: дуэль создана. Ожидаем принятия вызова.", true)
	h.notifyOpponent(r, d, oppID, "⚔️ Тебе бросили вызов в CLUTCH!")
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusCreated, card)
}

type acceptDuelRequest struct {
	OpponentTx *string `json:"opponent_tx"`
}

func (h *DuelHandler) Accept(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	var req acceptDuelRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	d, err := h.duels.Accept(r.Context(), duelID, userID, req.OpponentTx)
	if err != nil || d == nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_accept", "")
		return
	}
	_, _ = h.chat.Insert(r.Context(), d.ID, nil, "⚖️ Судья: дуэль активна. Удачи!", true)
	h.notifyOpponent(r, d, d.CreatorID, "✅ Вызов принят — дуэль началась!")
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DuelHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.Cancel(r.Context(), duelID, userID)
	if err != nil || d == nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_cancel", "")
		return
	}
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DuelHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || !service.UserInDuel(*d, userID) {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DuelHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	list, err := h.duels.ListForUser(r.Context(), userID, []models.DuelStatus{
		models.DuelPendingOpponent, models.DuelActive, models.DuelAwaitingClaim,
		models.DuelDisputed, models.DuelArbitrationUpload, models.DuelAIJudging,
		models.DuelAppealWindow, models.DuelHumanArbitration,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	cards, _ := h.cards.EnrichMany(r.Context(), list)
	httputil.JSON(w, http.StatusOK, cards)
}

func (h *DuelHandler) Claim(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || !service.UserInDuel(*d, userID) {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if d.Status != models.DuelActive {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}
	if err := h.duels.SetAwaitingClaim(r.Context(), duelID, userID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_claim", "")
		return
	}
	payload, _ := json.Marshal(map[string]string{"user_id": userID.String(), "action": "claim"})
	_ = h.duels.AddEvent(r.Context(), duelID, "claim", payload)
	_, _ = h.chat.Insert(r.Context(), duelID, &userID, "Заявил(а) победу — жду подтверждения соперника.", false)
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "claim_recorded"})
}

func (h *DuelHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || !service.UserInDuel(*d, userID) {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if d.Status != models.DuelAwaitingClaim {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}
	if d.ClaimedBy != nil && *d.ClaimedBy == userID {
		httputil.Error(w, http.StatusBadRequest, "claimer_cannot_confirm", "")
		return
	}
	other := service.OtherUserID(*d, userID)
	if other == nil {
		httputil.Error(w, http.StatusBadRequest, "invalid", "")
		return
	}
	winnerID := *other
	d, err = h.duels.Settle(r.Context(), duelID, winnerID, models.DuelMutualSettled)
	if err != nil || d == nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_settle", "")
		return
	}
	_, _ = h.chat.Insert(r.Context(), duelID, nil, "⚖️ Судья: победа подтверждена обеими сторонами. Банк распределён.", true)
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

type postMessageRequest struct {
	Body string `json:"body"`
}

func (h *DuelHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	var req postMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "body required")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || !service.UserInDuel(*d, userID) {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if d.Status != models.DuelActive && d.Status != models.DuelAwaitingClaim {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}
	msg, err := h.chat.Insert(r.Context(), duelID, &userID, strings.TrimSpace(req.Body), false)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	h.hub.Broadcast(duelID.String(), *msg)
	httputil.JSON(w, http.StatusCreated, msg)
}

func (h *DuelHandler) WS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	claims, err := auth.ParseAccessToken(h.cfg.JWTSecret, token)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	userID := claims.UserID
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || !service.UserInDuel(*d, userID) {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}

	conn, err := duelUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	roomID := duelID.String()
	h.hub.Join(roomID, conn)
	defer h.hub.Leave(roomID, conn)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var in struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(data, &in); err != nil || strings.TrimSpace(in.Body) == "" {
			continue
		}
		if d.Status != models.DuelActive && d.Status != models.DuelAwaitingClaim {
			continue
		}
		msg, err := h.chat.Insert(r.Context(), duelID, &userID, strings.TrimSpace(in.Body), false)
		if err != nil {
			continue
		}
		h.hub.Broadcast(roomID, *msg)
	}
}

func (h *DuelHandler) Messages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || !service.UserInDuel(*d, userID) {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	msgs, err := h.chat.List(r.Context(), duelID, 200)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	httputil.JSON(w, http.StatusOK, msgs)
}

type txSignatureRequest struct {
	Signature string `json:"signature"`
}

func (h *DuelHandler) TxCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	if h.chain == nil || !h.chain.HasProgram() {
		httputil.Error(w, http.StatusServiceUnavailable, "chain_unavailable", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || d.CreatorID != userID {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if d.CreatorTx != nil && *d.CreatorTx != "" {
		httputil.Error(w, http.StatusBadRequest, "already_submitted", "")
		return
	}
	wallet, err := h.userWallet(r, userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "no_wallet", err.Error())
		return
	}
	stake := uint64(d.StakeUSDEach)
	txB64, pda, err := h.chain.BuildCreateDuelTx(r.Context(), wallet, stake, d.DeadlineAt)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "build_tx_failed", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{
		"transaction":      txB64,
		"on_chain_duel_id": pda.String(),
	})
}

func (h *DuelHandler) TxConfirmCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	if h.chain == nil {
		httputil.Error(w, http.StatusServiceUnavailable, "chain_unavailable", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	var req txSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Signature == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "signature required")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || d.CreatorID != userID {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if err := h.chain.VerifyConfirmedTx(r.Context(), req.Signature); err != nil {
		httputil.Error(w, http.StatusBadRequest, "tx_not_confirmed", err.Error())
		return
	}
	wallet, err := h.userWallet(r, userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "no_wallet", err.Error())
		return
	}
	pda, _, _ := h.chain.DuelPDA(wallet)
	if err := h.duels.SetCreatorOnChain(r.Context(), duelID, req.Signature, pda.String()); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	d, _ = h.duels.GetByID(r.Context(), duelID)
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DuelHandler) TxAccept(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	if h.chain == nil || !h.chain.HasProgram() {
		httputil.Error(w, http.StatusServiceUnavailable, "chain_unavailable", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || d.OpponentID == nil || *d.OpponentID != userID {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if d.Status != models.DuelPendingOpponent {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}
	opponent, err := h.userWallet(r, userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "no_wallet", err.Error())
		return
	}
	creator, err := h.userWallet(r, d.CreatorID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "creator_wallet", "")
		return
	}
	txB64, err := h.chain.BuildAcceptDuelTx(r.Context(), opponent, creator)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "build_tx_failed", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"transaction": txB64})
}

func (h *DuelHandler) TxConfirmAccept(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	if h.chain == nil {
		httputil.Error(w, http.StatusServiceUnavailable, "chain_unavailable", "")
		return
	}
	duelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	var req txSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Signature == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "signature required")
		return
	}
	d, err := h.duels.GetByID(r.Context(), duelID)
	if err != nil || d == nil || d.OpponentID == nil || *d.OpponentID != userID {
		httputil.Error(w, http.StatusNotFound, "not_found", "")
		return
	}
	if err := h.chain.VerifyConfirmedTx(r.Context(), req.Signature); err != nil {
		httputil.Error(w, http.StatusBadRequest, "tx_not_confirmed", err.Error())
		return
	}
	_ = h.duels.SetOpponentOnChain(r.Context(), duelID, req.Signature)
	d, err = h.duels.Accept(r.Context(), duelID, userID, &req.Signature)
	if err != nil || d == nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_accept", "")
		return
	}
	_, _ = h.chat.Insert(r.Context(), d.ID, nil, "⚖️ Судья: дуэль активна. Удачи!", true)
	h.notifyOpponent(r, d, d.CreatorID, "✅ Вызов принят — дуэль началась!")
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DuelHandler) userWallet(r *http.Request, userID uuid.UUID) (solanago.PublicKey, error) {
	u, err := h.users.GetByID(r.Context(), userID)
	if err != nil || u == nil || u.WalletAddress == nil {
		return solanago.PublicKey{}, fmt.Errorf("wallet not linked")
	}
	return solanago.PublicKeyFromBase58(*u.WalletAddress)
}

func (h *DuelHandler) notifyOpponent(r *http.Request, d *models.Duel, oppID uuid.UUID, text string) {
	u, err := h.users.GetByID(r.Context(), oppID)
	if err != nil || u == nil {
		return
	}
	_ = h.notify.SendText(u.TelegramID, text, h.cfg.MiniAppPublicURL+"/duel/"+d.ID.String())
}

func (h *DuelHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/active", h.ListActive)
	r.Route("/{id}", func(dr chi.Router) {
		dr.Get("/", h.Get)
		dr.Post("/accept", h.Accept)
		dr.Post("/cancel", h.Cancel)
		dr.Post("/claim", h.Claim)
		dr.Post("/confirm", h.Confirm)
		dr.Get("/messages", h.Messages)
		dr.Post("/messages", h.PostMessage)
		dr.Get("/ws", h.WS)
		dr.Get("/tx/create", h.TxCreate)
		dr.Post("/tx/create", h.TxConfirmCreate)
		dr.Get("/tx/accept", h.TxAccept)
		dr.Post("/tx/accept", h.TxConfirmAccept)
	})
	return r
}
