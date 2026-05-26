# CLUTCH

Telegram Mini App для дуэлей со ставками в Solana.  
Архитектура: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) · Phase 0: [docs/PHASE_0_PLAN.md](docs/PHASE_0_PLAN.md)

**Бот:** [@clutch_game_bot](https://t.me/clutch_game_bot)  
**Репозиторий:** [github.com/GalahadKingsman/clutch](https://github.com/GalahadKingsman/clutch)

---

## Стек

| Часть | Технология |
|-------|------------|
| Mini App | React, Vite, Tailwind |
| API + Bot | Go 1.25+ |
| DB | PostgreSQL 16, Redis 7 |
| Chain | Solana devnet, Anchor 0.30 |
| Deploy | Docker Compose + nginx |

---

## Быстрый старт на VPS

### 1. Зависимости на сервере

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y docker.io docker-compose-plugin git make

# Go 1.25 (если нужна локальная сборка без Docker)
# https://go.dev/dl/

# Node 20+ (для сборки miniapp на VPS)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
```

### 2. Клонирование и конфиг

```bash
git clone https://github.com/GalahadKingsman/clutch.git
cd clutch
cp .env.example .env
nano .env   # заполни TELEGRAM_BOT_TOKEN, JWT_SECRET, PUBLIC_URL, пароли
```

**Обязательные переменные:**

```env
TELEGRAM_BOT_TOKEN=...
JWT_SECRET=длинная-случайная-строка
POSTGRES_PASSWORD=...
PUBLIC_URL=https://твой-домен.com
API_PUBLIC_URL=https://твой-домен.com
MINIAPP_PUBLIC_URL=https://твой-домен.com
CORS_ORIGINS=https://твой-домен.com
TELEGRAM_WEBHOOK_PUBLIC_URL=https://твой-домен.com/telegram/webhook
TELEGRAM_WEBHOOK_SECRET=случайный-секрет
```

### 3. WalletConnect (привязка кошелька в Mini App)

1. [cloud.reown.com](https://cloud.reown.com) → Project ID  
2. В `.env` на VPS: `VITE_WALLETCONNECT_PROJECT_ID=...`, `VITE_APP_URL=https://clutch-duel.ru`  
3. В Reown → Allowed origins: `https://clutch-duel.ru`  
4. Подробнее: [docs/WALLETCONNECT.md](docs/WALLETCONNECT.md)

### 4. BotFather (делаешь ты)

1. [@BotFather](https://t.me/BotFather) → твой бот `@clutch_game_bot`
2. `/setdomain` → домен Mini App (тот же что `PUBLIC_URL`)
3. `/setmenubutton` → Web App URL = `MINIAPP_PUBLIC_URL`
4. Webhook секрет = `TELEGRAM_WEBHOOK_SECRET` (тот же в заголовке Telegram)

### 5. Деплой

```bash
# Полный деплой (miniapp build + backend + nginx)
make deploy-all

# Если ошибка «npm ci» / нет package-lock.json:
# cd apps/miniapp && npm install && npm run build && cd ../..
# затем docker compose ... (см. Makefile)

# Только пересобрать бэкенд
make redeploy-backend

# Только пересобрать Mini App (после правок UI)
make redeploy-miniapp
```

**Белый экран в Telegram, а в браузере «Открой через Telegram»?** → **нет HTTPS**. Telegram Mini App работает только по `https://`. См. [docs/HTTPS_TELEGRAM.md](docs/HTTPS_TELEGRAM.md).

**Белый экран везде?** Почти всегда nginx без собранного фронта. На VPS:

```bash
make redeploy-backend   # пересоберёт api + nginx (miniapp внутри Docker)
# проверка:
curl -sI https://YOUR_DOMAIN/ | head -3
curl -s https://YOUR_DOMAIN/ | head -5   # должен быть index.html с /assets/*.js
curl -s https://YOUR_DOMAIN/api/v1/../health  # → /health через nginx: GET /health на api
curl -s https://YOUR_DOMAIN/health
docker compose logs api --tail 30
```

Nginx по умолчанию слушает **80**. HTTPS: Certbot на VPS + [docs/CERTBOT_VPS.md](docs/CERTBOT_VPS.md) или Cloudflare — [docs/HTTPS_TELEGRAM.md](docs/HTTPS_TELEGRAM.md).

### 5. Solana program (devnet)

На VPS или локально с установленным [Anchor](https://www.anchor-lang.com/docs/installation):

```bash
solana config set --url devnet
solana airdrop 2
anchor build
anchor deploy --provider.cluster devnet
# Program ID → CLUTCH_PROGRAM_ID в .env
```

---

## Локальная разработка

```bash
make dev-up          # postgres + redis
cp .env.example .env # DATABASE_URL=postgres://clutch:change-me@localhost:5432/clutch?sslmode=disable

go run ./cmd/migrate up
go run ./cmd/api

cd apps/miniapp && npm ci && npm run dev
```

Mini App без Telegram: в браузере покажет ошибку «Открой через Telegram» — это ожидаемо.

---

## Makefile

| Команда | Действие |
|---------|----------|
| `make deploy-all` | Собрать miniapp + поднять весь docker stack |
| `make deploy-backend` | API, bot, postgres, redis, nginx (без rebuild miniapp) |
| `make redeploy-backend` | Rebuild + restart api & bot |
| `make deploy-miniapp` | `vite build` + restart nginx |
| `make redeploy-miniapp` | То же |
| `make migrate-up` | Прогнать SQL миграции |

---

## Структура

```text
cmd/api/          # REST API
cmd/bot/          # Telegram webhook bot
cmd/migrate/      # SQL migrations
apps/miniapp/     # Telegram Mini App
programs/         # Anchor escrow
deploy/           # Dockerfiles, nginx
migrations/       # SQL
```

---

## Phase 0 status

- [x] Monorepo scaffold
- [x] Go API: TG auth, JWT, wallet link (SIWS)
- [x] Bot: /start + webhook
- [x] Mini App: Wallet Gate + feed stub
- [x] Docker Compose + Makefile deploy
- [x] Anchor program skeleton (create/accept/cancel)
- [ ] Devnet deploy program (на VPS с Anchor CLI)
- [ ] Реальный домен + BotFather
