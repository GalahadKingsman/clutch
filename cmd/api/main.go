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
	"github.com/GalahadKingsman/clutch/internal/service"
	"github.com/GalahadKingsman/clutch/internal/solana"
	"github.com/GalahadKingsman/clutch/internal/storage"
	"github.com/GalahadKingsman/clutch/internal/telegram"
	"github.com/GalahadKingsman/clutch/internal/ws"
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
	friends := repository.NewFriendRepository(pool)
	duels := repository.NewDuelRepository(pool)
	proofs := repository.NewProofRepository(pool)
	verdicts := repository.NewVerdictRepository(pool)
	chat := repository.NewChatRepository(pool)
	cards := service.NewDuelCardService(users)
	notify := telegram.NewNotifier(cfg.TelegramBotToken)
	hub := ws.NewHub()
	priceH := handlers.NewPriceHandler()

	chain, err := solana.NewClient(cfg.SolanaRPCURL, cfg.ClutchProgramID)
	if err != nil {
		log.Printf("warn: solana client: %v", err)
	}
	if chain != nil && chain.HasProgram() {
		log.Printf("clutch on-chain program: %s", cfg.ClutchProgramID)
	}

	authH := handlers.NewAuthHandler(cfg, users)
	socialH := handlers.NewSocialHandler(cfg, friends)
	feedH := handlers.NewFeedHandler(duels, cards)
	walletH := handlers.NewWalletHandler(cfg, users, chain)
	store, err := storage.NewLocalStore(cfg.UploadDir)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	duelH := handlers.NewDuelHandler(cfg, duels, friends, users, chat, cards, notify, hub, chain)
	disputeH := handlers.NewDisputeHandler(cfg, duels, proofs, verdicts, users, chat, cards, notify, store)
	aiH := handlers.NewAIHandler()

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
		api.Get("/auth/wallet/phantom-bridge/session", authH.PhantomBridgeSession)
		api.Post("/auth/wallet/phantom-bridge/link", authH.PhantomBridgeLink)

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.Authenticate(cfg.JWTSecret))
			protected.Get("/auth/me", authH.Me)
			protected.Post("/auth/wallet/nonce", authH.WalletNonce)
			protected.Post("/auth/wallet/link", authH.WalletLink)
			protected.Post("/auth/wallet/phantom-bridge", authH.PhantomBridgePrepare)

			protected.Group(func(locked chi.Router) {
				locked.Use(middleware.RequireWallet(users))
				locked.Get("/feed/friends", feedH.Friends)
				locked.Get("/prices", priceH.Get)
				locked.Get("/wallet/balances", walletH.Balances)
				locked.Get("/users/search", socialH.SearchUsers)
				locked.Mount("/friends", socialH.Routes())
				locked.Mount("/duels", duelH.Routes())
				locked.Route("/duels/{id}", func(dr chi.Router) {
					dr.Mount("/", disputeH.Routes())
				})
				locked.Post("/ai/clarify-condition", aiH.ClarifyCondition)
			})
		})
		api.Get("/files/*", disputeH.ServeFile)
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
