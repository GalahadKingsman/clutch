package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/google/uuid"
)

type JudgeInput struct {
	ConditionText string
	SideCreator   string
	SideOpponent  string
	CreatorID     uuid.UUID
	OpponentID    uuid.UUID
	ClaimedBy     *uuid.UUID
	Proofs        []models.Proof
}

type JudgeOutput struct {
	WinnerID    uuid.UUID
	Reasoning   string
	Confidence  float64
	EvidenceIDs []uuid.UUID
	VerdictHash string
}

type Service struct {
	apiKey string
	model  string
	client *http.Client
}

func NewService() *Service {
	return &Service{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		model:  getenv("OPENAI_MODEL", "gpt-4o-mini"),
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (s *Service) Clarify(ctx context.Context, condition, sideCreator, sideOpponent string) (*models.ClarifyResponse, error) {
	if s.apiKey != "" {
		prompt := fmt.Sprintf(
			"Условие дуэли между друзьями: %q. Сторона A: %s. Сторона B: %s. Верни JSON: normalized_condition, win_criterion, tips (кратко на русском).",
			condition, sideCreator, sideOpponent,
		)
		raw, err := s.chatCompletion(ctx, prompt)
		if err == nil {
			var out models.ClarifyResponse
			if json.Unmarshal([]byte(extractJSON(raw)), &out) == nil && out.NormalizedCondition != "" {
				return &out, nil
			}
		}
	}
	return &models.ClarifyResponse{
		NormalizedCondition: condition,
		WinCriterion:        "Победитель определяется по фактическому исходу события из условия.",
		Tips:                "Укажи объективный критерий (счёт, дата, источник) и загрузи скрин/фото как пруф.",
	}, nil
}

func (s *Service) Judge(ctx context.Context, in JudgeInput) (*JudgeOutput, error) {
	if s.apiKey != "" {
		out, err := s.judgeOpenAI(ctx, in)
		if err == nil {
			return out, nil
		}
	}
	return s.judgeHeuristic(in), nil
}

func (s *Service) judgeHeuristic(in JudgeInput) *JudgeOutput {
	winner := in.CreatorID
	reason := "Эвристика CLUTCH: по умолчанию в пользу создателя дуэли. Загрузите пруфы и используйте OpenAI для точного вердикта."

	if in.ClaimedBy != nil {
		winner = *in.ClaimedBy
		reason = "Эвристика: победитель совпадает с тем, кто заявил победу (до AI-анализа пруфов)."
	}

	// Больше пруфов у стороны → небольшой перевес
	creatorProofs, opponentProofs := 0, 0
	var refs []uuid.UUID
	for _, p := range in.Proofs {
		refs = append(refs, p.ID)
		if p.UserID == in.CreatorID {
			creatorProofs++
		} else {
			opponentProofs++
		}
	}
	if opponentProofs > creatorProofs {
		winner = in.OpponentID
		reason = fmt.Sprintf("Эвристика: у соперника больше доказательств (%d vs %d).", opponentProofs, creatorProofs)
	} else if creatorProofs > opponentProofs {
		winner = in.CreatorID
		reason = fmt.Sprintf("Эвристика: у создателя больше доказательств (%d vs %d).", creatorProofs, opponentProofs)
	}

	conf := 0.72
	if len(in.Proofs) == 0 {
		conf = 0.55
		reason += " Пруфы не загружены — низкая уверенность."
	}

	hash := verdictHash(winner, reason, refs)
	return &JudgeOutput{
		WinnerID:    winner,
		Reasoning:   reason,
		Confidence:  conf,
		EvidenceIDs: refs,
		VerdictHash: hash,
	}
}

func (s *Service) judgeOpenAI(ctx context.Context, in JudgeInput) (*JudgeOutput, error) {
	var proofLines []string
	var refs []uuid.UUID
	for _, p := range in.Proofs {
		refs = append(refs, p.ID)
		cap := ""
		if p.Caption != nil {
			cap = *p.Caption
		}
		proofLines = append(proofLines, fmt.Sprintf("- user=%s type=%s caption=%s id=%s", p.UserID, p.ProofType, cap, p.ID))
	}
	claimer := "unknown"
	if in.ClaimedBy != nil {
		claimer = in.ClaimedBy.String()
	}
	prompt := fmt.Sprintf(`Ты судья CLUTCH. Определи победителя дуэли.
Условие: %q
Сторона создателя (id %s): %s
Сторона оппонента (id %s): %s
Кто заявил победу: %s
Доказательства:
%s
Верни ТОЛЬКО JSON: {"winner_user_id":"uuid","confidence":0.0-1.0,"reasoning":"..."}`,
		in.ConditionText, in.CreatorID, in.SideCreator, in.OpponentID, in.SideOpponent, claimer, strings.Join(proofLines, "\n"))

	raw, err := s.chatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		WinnerUserID string  `json:"winner_user_id"`
		Confidence   float64 `json:"confidence"`
		Reasoning    string  `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &parsed); err != nil {
		return nil, err
	}
	winner, err := uuid.Parse(parsed.WinnerUserID)
	if err != nil {
		if winner == in.CreatorID || parsed.WinnerUserID == in.CreatorID.String() {
			winner = in.CreatorID
		} else {
			winner = in.OpponentID
		}
	}
	if parsed.Confidence <= 0 || parsed.Confidence > 1 {
		parsed.Confidence = 0.85
	}
	hash := verdictHash(winner, parsed.Reasoning, refs)
	return &JudgeOutput{
		WinnerID:    winner,
		Reasoning:   parsed.Reasoning,
		Confidence:  parsed.Confidence,
		EvidenceIDs: refs,
		VerdictHash: hash,
	}, nil
}

func (s *Service) chatCompletion(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "system", "content": "Ты арбитр споров CLUTCH. Отвечай кратко, по делу."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("openai status %d", res.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Choices) == 0 {
		return "", fmt.Errorf("openai parse error")
	}
	return out.Choices[0].Message.Content, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}

func verdictHash(winner uuid.UUID, reason string, refs []uuid.UUID) string {
	h := sha256.New()
	h.Write([]byte(winner.String()))
	h.Write([]byte(reason))
	for _, id := range refs {
		h.Write([]byte(id.String()))
	}
	return hex.EncodeToString(h.Sum(nil))
}
