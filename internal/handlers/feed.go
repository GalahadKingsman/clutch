package handlers

import (
	"net/http"

	"github.com/GalahadKingsman/clutch/internal/httputil"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/GalahadKingsman/clutch/internal/service"
)

type FeedHandler struct {
	duels  *repository.DuelRepository
	cards  *service.DuelCardService
}

func NewFeedHandler(duels *repository.DuelRepository, cards *service.DuelCardService) *FeedHandler {
	return &FeedHandler{duels: duels, cards: cards}
}

func (h *FeedHandler) Friends(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	incoming, err := h.duels.ListIncoming(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}
	active, err := h.duels.ListForUser(r.Context(), userID, []models.DuelStatus{
		models.DuelActive, models.DuelAwaitingClaim,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal", "")
		return
	}

	inCards, _ := h.cards.EnrichMany(r.Context(), incoming)
	actCards, _ := h.cards.EnrichMany(r.Context(), active)

	httputil.JSON(w, http.StatusOK, models.FeedResponse{
		IncomingChallenges: inCards,
		ActiveDuels:        actCards,
		Activity:           []models.FeedActivity{},
	})
}
