package handlers

import (
	"net/http"

	"github.com/GalahadKingsman/clutch/internal/httputil"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "clutch-api"})
}
