package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/GalahadKingsman/clutch/internal/httputil"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/go-chi/chi/v5"
)

type SocialHandler struct {
	cfg     *config.Config
	friends *repository.FriendRepository
}

func NewSocialHandler(cfg *config.Config, friends *repository.FriendRepository) *SocialHandler {
	return &SocialHandler{cfg: cfg, friends: friends}
}

func (h *SocialHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	list, err := h.friends.ListFriends(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	httputil.JSON(w, http.StatusOK, list)
}

func (h *SocialHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		httputil.JSON(w, http.StatusOK, []any{})
		return
	}
	users, err := h.friends.SearchUsers(r.Context(), q, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	out := make([]any, 0, len(users))
	for _, u := range users {
		out = append(out, u.ToResponse())
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *SocialHandler) SearchFriends(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	q := r.URL.Query().Get("q")
	list, err := h.friends.SearchFriends(r.Context(), userID, q)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	httputil.JSON(w, http.StatusOK, list)
}

func (h *SocialHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	code, err := randomCode(8)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	expires := time.Now().Add(7 * 24 * time.Hour)
	if err := h.friends.CreateInviteCode(r.Context(), code, userID, expires); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	link := "https://t.me/" + h.cfg.TelegramBotUser + "?startapp=invite_" + code
	httputil.JSON(w, http.StatusOK, map[string]string{
		"code": code,
		"link": link,
	})
}

type acceptInviteRequest struct {
	Code         string  `json:"code"`
	ContactAlias *string `json:"contact_alias"`
}

func (h *SocialHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "code required")
		return
	}
	// strip prefix from startapp
	code := req.Code
	if len(code) > 7 && code[:7] == "invite_" {
		code = code[7:]
	}
	inviterID, err := h.friends.AcceptInvite(r.Context(), code, userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_invite", err.Error())
		return
	}
	_ = inviterID
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func randomCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *SocialHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListFriends)
	r.Get("/search", h.SearchFriends)
	r.Post("/invite", h.CreateInvite)
	r.Post("/accept", h.AcceptInvite)
	return r
}
