package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/GalahadKingsman/clutch/internal/httputil"
)

type PriceHandler struct {
	mu       sync.Mutex
	cached   map[string]float64
	cachedAt time.Time
	ttl      time.Duration
}

func NewPriceHandler() *PriceHandler {
	return &PriceHandler{
		cached: make(map[string]float64),
		ttl:    60 * time.Second,
	}
}

func (h *PriceHandler) Get(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "SOL"
	}

	h.mu.Lock()
	if time.Since(h.cachedAt) < h.ttl && len(h.cached) > 0 {
		out := h.cached
		h.mu.Unlock()
		httputil.JSON(w, http.StatusOK, out)
		return
	}
	h.mu.Unlock()

	prices, err := fetchJupiterPrices()
	if err != nil || len(prices) == 0 {
		prices = map[string]float64{
			"SOL":  145.0,
			"USDC": 1.0,
		}
	}

	h.mu.Lock()
	h.cached = prices
	h.cachedAt = time.Now()
	h.mu.Unlock()

	httputil.JSON(w, http.StatusOK, prices)
}

func fetchJupiterPrices() (map[string]float64, error) {
	mints := map[string]string{
		"SOL":  "So11111111111111111111111111111111111111112",
		"USDC": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
	}
	reqBody, _ := json.Marshal(map[string][]string{
		"ids": {mints["SOL"], mints["USDC"]},
	})
	resp, err := http.Post(
		"https://api.jup.ag/price/v2",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]struct {
		USDPrice float64 `json:"usdPrice"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	out := map[string]float64{"USDC": 1.0}
	if p, ok := raw[mints["SOL"]]; ok && p.USDPrice > 0 {
		out["SOL"] = p.USDPrice
	}
	return out, nil
}
