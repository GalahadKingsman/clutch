package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/GalahadKingsman/clutch/internal/ai"
	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/GalahadKingsman/clutch/internal/httputil"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/GalahadKingsman/clutch/internal/service"
	"github.com/GalahadKingsman/clutch/internal/storage"
	"github.com/GalahadKingsman/clutch/internal/telegram"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const appealWindowDuration = 5 * time.Minute

type DisputeHandler struct {
	cfg      *config.Config
	duels    *repository.DuelRepository
	proofs   *repository.ProofRepository
	verdicts *repository.VerdictRepository
	users    *repository.UserRepository
	chat     *repository.ChatRepository
	cards    *service.DuelCardService
	notify   *telegram.Notifier
	store    *storage.LocalStore
	ai       *ai.Service
}

func NewDisputeHandler(
	cfg *config.Config,
	duels *repository.DuelRepository,
	proofs *repository.ProofRepository,
	verdicts *repository.VerdictRepository,
	users *repository.UserRepository,
	chat *repository.ChatRepository,
	cards *service.DuelCardService,
	notify *telegram.Notifier,
	store *storage.LocalStore,
) *DisputeHandler {
	return &DisputeHandler{
		cfg: cfg, duels: duels, proofs: proofs, verdicts: verdicts,
		users: users, chat: chat, cards: cards, notify: notify,
		store: store, ai: ai.NewService(),
	}
}

func (h *DisputeHandler) Dispute(w http.ResponseWriter, r *http.Request) {
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
		httputil.Error(w, http.StatusBadRequest, "claimer_cannot_dispute", "")
		return
	}
	d, err = h.duels.OpenDispute(r.Context(), duelID, userID)
	if err != nil || d == nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_dispute", "")
		return
	}
	_, _ = h.chat.Insert(r.Context(), duelID, nil, "⚖️ Судья: открыт спор. Загрузите доказательства.", true)
	h.notifyDispute(r, d)
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DisputeHandler) UploadProof(w http.ResponseWriter, r *http.Request) {
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
	if d.Status != models.DuelDisputed && d.Status != models.DuelArbitrationUpload {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}
	if d.Status == models.DuelDisputed {
		_, _ = h.duels.SetStatus(r.Context(), duelID, models.DuelArbitrationUpload)
	}

	if err := r.ParseMultipartForm(12 << 20); err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "multipart required")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "file required")
		return
	}
	defer file.Close()

	rel, hash, err := h.store.SaveProof(duelID, userID, hdr.Filename, file)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "storage_error", "")
		return
	}
	caption := r.FormValue("caption")
	var cap *string
	if caption != "" {
		cap = &caption
	}
	p, err := h.proofs.Insert(r.Context(), repository.InsertProofInput{
		DuelID: duelID, UserID: userID, ProofType: "image",
		StoragePath: rel, Caption: cap, ContentHash: hash,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	p.URL = h.cfg.APIPublicURL + h.store.PublicPath(rel)
	httputil.JSON(w, http.StatusCreated, p)
}

func (h *DisputeHandler) ListProofs(w http.ResponseWriter, r *http.Request) {
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
	list, err := h.proofs.ListByDuel(r.Context(), duelID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	for i := range list {
		list[i].URL = h.cfg.APIPublicURL + h.store.PublicPath(list[i].StoragePath)
	}
	httputil.JSON(w, http.StatusOK, list)
}

func (h *DisputeHandler) RunJudge(w http.ResponseWriter, r *http.Request) {
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
	if d.Status != models.DuelDisputed && d.Status != models.DuelArbitrationUpload {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}

	_, _ = h.duels.SetStatus(r.Context(), duelID, models.DuelAIJudging)
	proofs, _ := h.proofs.ListByDuel(r.Context(), duelID)
	if len(proofs) == 0 {
		httputil.Error(w, http.StatusBadRequest, "proofs_required", "upload at least one proof")
		return
	}

	oppID := *d.OpponentID
	out, err := h.ai.Judge(r.Context(), ai.JudgeInput{
		ConditionText: d.ConditionText,
		SideCreator:   d.SideCreator,
		SideOpponent:  d.SideOpponent,
		CreatorID:     d.CreatorID,
		OpponentID:    oppID,
		ClaimedBy:     d.ClaimedBy,
		Proofs:        proofs,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}

	v, err := h.verdicts.Insert(r.Context(), repository.InsertVerdictInput{
		DuelID: duelID, WinnerID: out.WinnerID, Reasoning: out.Reasoning,
		Confidence: out.Confidence, EvidenceRefs: out.EvidenceIDs, VerdictHash: out.VerdictHash,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}

	ends := time.Now().Add(appealWindowDuration)
	d, err = h.duels.SetAppealWindow(r.Context(), duelID, v.ID, ends)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}

	msg := fmt.Sprintf("⚖️ Судья: вердикт ИИ. Победитель определён. Апелляция возможна %d мин.", int(appealWindowDuration.Minutes()))
	_, _ = h.chat.Insert(r.Context(), duelID, nil, msg, true)
	h.notifyVerdict(r, d, v, ends)

	resp := h.buildVerdictResponse(d, v, userID)
	httputil.JSON(w, http.StatusOK, resp)
}

func (h *DisputeHandler) GetVerdict(w http.ResponseWriter, r *http.Request) {
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
	h.maybeFinalizeAppealWindow(r, d)

	d, _ = h.duels.GetByID(r.Context(), duelID)
	v, err := h.verdicts.GetByDuel(r.Context(), duelID)
	if err != nil || v == nil {
		httputil.Error(w, http.StatusNotFound, "no_verdict", "")
		return
	}
	httputil.JSON(w, http.StatusOK, h.buildVerdictResponse(d, v, userID))
}

func (h *DisputeHandler) FinalizeVerdict(w http.ResponseWriter, r *http.Request) {
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
	v, err := h.verdicts.GetByDuel(r.Context(), duelID)
	if err != nil || v == nil {
		httputil.Error(w, http.StatusNotFound, "no_verdict", "")
		return
	}
	if d.Status != models.DuelAppealWindow {
		httputil.Error(w, http.StatusBadRequest, "invalid_status", "")
		return
	}
	d, err = h.duels.Settle(r.Context(), duelID, v.WinnerID, models.DuelSettled)
	if err != nil || d == nil {
		httputil.Error(w, http.StatusBadRequest, "cannot_settle", "")
		return
	}
	_, _ = h.chat.Insert(r.Context(), duelID, nil, "⚖️ Судья: вердикт финализирован. Банк распределён.", true)
	h.notifySettled(r, d, v.WinnerID)
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, card)
}

func (h *DisputeHandler) Appeal(w http.ResponseWriter, r *http.Request) {
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
	v, err := h.verdicts.GetByDuel(r.Context(), duelID)
	if err != nil || v == nil {
		httputil.Error(w, http.StatusNotFound, "no_verdict", "")
		return
	}
	if d.Status != models.DuelAppealWindow {
		httputil.Error(w, http.StatusBadRequest, "appeal_closed", "")
		return
	}
	if d.AppealWindowEndsAt != nil && time.Now().After(*d.AppealWindowEndsAt) {
		httputil.Error(w, http.StatusBadRequest, "appeal_expired", "")
		return
	}
	if v.WinnerID == userID {
		httputil.Error(w, http.StatusBadRequest, "winner_cannot_appeal", "")
		return
	}

	fee := d.BankUSD * 0.05
	sla := time.Now().Add(24 * time.Hour)
	appeal, err := h.verdicts.InsertAppeal(r.Context(), duelID, userID, fee, sla)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	d, err = h.duels.SetHumanAppeal(r.Context(), duelID, appeal.ID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	_, _ = h.chat.Insert(r.Context(), duelID, nil, "⚖️ Судья: запрошена апелляция к человеку-арбитру (24ч).", true)
	h.notifyAppeal(r, d, appeal)
	card, _ := h.cards.Enrich(r.Context(), *d)
	httputil.JSON(w, http.StatusOK, map[string]any{"duel": card, "appeal": appeal})
}

func (h *DisputeHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	rel := chi.URLParam(r, "*")
	rel = strings.TrimPrefix(rel, "/")
	f, err := h.store.Open(rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	w.Header().Set("Cache-Control", "private, max-age=3600")
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	_, _ = io.Copy(w, f)
}

func (h *DisputeHandler) maybeFinalizeAppealWindow(r *http.Request, d *models.Duel) {
	if d == nil || d.Status != models.DuelAppealWindow || d.AppealWindowEndsAt == nil {
		return
	}
	if time.Now().Before(*d.AppealWindowEndsAt) {
		return
	}
	v, _ := h.verdicts.GetByDuel(r.Context(), d.ID)
	if v == nil {
		return
	}
	settled, _ := h.duels.Settle(r.Context(), d.ID, v.WinnerID, models.DuelSettled)
	if settled != nil {
		_, _ = h.chat.Insert(r.Context(), d.ID, nil, "⚖️ Судья: окно апелляции истекло — вердикт в силе.", true)
		h.notifySettled(r, settled, v.WinnerID)
	}
}

func (h *DisputeHandler) buildVerdictResponse(d *models.Duel, v *models.AIVerdict, viewer uuid.UUID) models.AIVerdict {
	resp := *v
	resp.IsWinner = v.WinnerID == viewer
	if d != nil {
		resp.AppealEndsAt = d.AppealWindowEndsAt
		if d.Status == models.DuelAppealWindow && d.AppealWindowEndsAt != nil && time.Now().Before(*d.AppealWindowEndsAt) && !resp.IsWinner {
			resp.CanAppeal = true
		}
	}
	return resp
}

func (h *DisputeHandler) notifyDispute(r *http.Request, d *models.Duel) {
	if d.ClaimedBy == nil {
		return
	}
	u, _ := h.users.GetByID(r.Context(), *d.ClaimedBy)
	if u != nil {
		_ = h.notify.SendText(u.TelegramID, "⚔️ Соперник оспорил исход. Загрузи пруфы в CLUTCH.",
			h.cfg.MiniAppPublicURL+"/duel/"+d.ID.String()+"/arbitration")
	}
}

func (h *DisputeHandler) notifyVerdict(r *http.Request, d *models.Duel, v *models.AIVerdict, ends time.Time) {
	for _, uid := range []uuid.UUID{d.CreatorID, *d.OpponentID} {
		u, _ := h.users.GetByID(r.Context(), uid)
		if u == nil {
			continue
		}
		text := "⚖️ Вердикт ИИ готов."
		if uid == v.WinnerID {
			text = "🏆 Вердикт ИИ: победа твоя!"
		} else {
			text = fmt.Sprintf("⚖️ Вердикт ИИ: победа соперника. Апелляция до %s.", ends.Format("15:04"))
		}
		_ = h.notify.SendText(u.TelegramID, text, h.cfg.MiniAppPublicURL+"/duel/"+d.ID.String()+"/verdict")
	}
}

func (h *DisputeHandler) notifySettled(r *http.Request, d *models.Duel, winnerID uuid.UUID) {
	for _, uid := range []uuid.UUID{d.CreatorID, *d.OpponentID} {
		u, _ := h.users.GetByID(r.Context(), uid)
		if u == nil {
			continue
		}
		text := "⚖️ Дуэль завершена."
		if uid == winnerID {
			text = "🏆 Дуэль завершена — победа твоя!"
		}
		_ = h.notify.SendText(u.TelegramID, text, h.cfg.MiniAppPublicURL+"/duel/"+d.ID.String())
	}
}

func (h *DisputeHandler) notifyAppeal(r *http.Request, d *models.Duel, appeal *models.HumanAppeal) {
	if h.cfg.TelegramArbiterChatID == 0 {
		return
	}
	_ = h.notify.SendText(h.cfg.TelegramArbiterChatID,
		fmt.Sprintf("📋 Апелляция CLUTCH\nДуэль: %s\nАпеллянт: %s\nБанк: $%.0f",
			d.ID.String(), appeal.AppellantID.String(), d.BankUSD),
		h.cfg.MiniAppPublicURL+"/duel/"+d.ID.String())
}

func (h *DisputeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/dispute", h.Dispute)
	r.Get("/proofs", h.ListProofs)
	r.Post("/proofs", h.UploadProof)
	r.Post("/judge", h.RunJudge)
	r.Get("/verdict", h.GetVerdict)
	r.Post("/verdict/finalize", h.FinalizeVerdict)
	r.Post("/appeal", h.Appeal)
	return r
}
