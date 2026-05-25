package main

import (
	"log"
	"os"

	"github.com/GalahadKingsman/clutch/internal/config"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	webhookPublic := os.Getenv("TELEGRAM_WEBHOOK_PUBLIC_URL")
	if webhookPublic == "" {
		log.Fatal("TELEGRAM_WEBHOOK_PUBLIC_URL is required for production bot (e.g. https://api.example.com/telegram/webhook)")
	}

	miniAppURL := cfg.MiniAppPublicURL

	pref := tele.Settings{
		Token: cfg.TelegramBotToken,
		Poller: &tele.Webhook{
			Listen:      ":" + cfg.BotWebhookPort,
			SecretToken: cfg.TelegramWebhookSec,
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: webhookPublic,
			},
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}

	b.Handle("/start", func(c tele.Context) error {
		payload := c.Message().Payload
		appURL := miniAppURL
		msg := "⚔️ Добро пожаловать в CLUTCH!\n\nСпорь с друзьями на крипту — Судья следит за честностью."
		if payload != "" {
			if len(payload) > 7 && payload[:7] == "invite_" {
				msg = "👋 Тебя пригласили в CLUTCH!\n\nОткрой приложение — дружба добавится автоматически."
				appURL = miniAppURL + "?startapp=" + payload
			} else {
				msg += "\n\nКод: " + payload
			}
		}
		return c.Send(msg, &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
			{{Text: "Открыть CLUTCH", WebApp: &tele.WebApp{URL: appURL}}},
		}})
	})

	log.Printf("clutch-bot starting webhook on :%s → %s", cfg.BotWebhookPort, webhookPublic)
	b.Start()
}
