package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/GalahadKingsman/clutch/internal/ai"
	"github.com/GalahadKingsman/clutch/internal/httputil"
)

type AIHandler struct {
	ai *ai.Service
}

func NewAIHandler() *AIHandler {
	return &AIHandler{ai: ai.NewService()}
}

type clarifyRequest struct {
	ConditionText string `json:"condition_text"`
	SideCreator   string `json:"side_creator"`
	SideOpponent  string `json:"side_opponent"`
}

func (h *AIHandler) ClarifyCondition(w http.ResponseWriter, r *http.Request) {
	var req clarifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ConditionText == "" {
		httputil.Error(w, http.StatusBadRequest, "bad_request", "")
		return
	}
	out, err := h.ai.Clarify(r.Context(), req.ConditionText, req.SideCreator, req.SideOpponent)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, out)
}
