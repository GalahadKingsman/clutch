package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/GalahadKingsman/clutch/internal/database"
	"github.com/GalahadKingsman/clutch/internal/handlers"
	"github.com/GalahadKingsman/clutch/internal/middleware"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	users := repository.NewUserRepository(pool)
	authH := handlers.NewAuthHandler(cfg, users)

	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, chimw.Logger, chimw.Recoverer, chimw.Timeout(60*time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", handlers.Health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/telegram", authH.Telegram)

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.Authenticate(cfg.JWTSecret))
			protected.Get("/auth/me", authH.Me)
			protected.Post("/auth/wallet/nonce", authH.WalletNonce)
			protected.Post("/auth/wallet/link", authH.WalletLink)

			// Phase 1+ routes go behind RequireWallet
			protected.Group(func(locked chi.Router) {
				locked.Use(middleware.RequireWallet(users))
				locked.Get("/feed/friends", func(w http.ResponseWriter, r *http.Request) {
					handlers.Health(w, r) // stub until Phase 1
				})
			})
		})
	})

	addr := ":" + cfg.APIPort
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Printf("clutch-api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
