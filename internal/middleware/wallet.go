package middleware

import (
	"net/http"

	"github.com/GalahadKingsman/clutch/internal/repository"
)

// RequireWallet blocks API access until user linked a Solana wallet.
func RequireWallet(users *repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			user, err := users.GetByID(r.Context(), userID)
			if err != nil || user == nil {
				http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
				return
			}
			if !user.WalletLinked() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"wallet_required","message":"Link a Solana wallet first"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
