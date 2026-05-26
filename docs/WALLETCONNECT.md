# WalletConnect v2 (Solana) — Reown AppKit

Привязка кошелька и on-chain tx в Mini App идут через **Reown AppKit** (WalletConnect v2). Phantom, Trust и др. — в модальном окне выбора кошелька.

## 1. Project ID

1. Зарегистрируй проект на [https://cloud.reown.com](https://cloud.reown.com) (бывший WalletConnect Cloud).
2. Скопируй **Project ID**.

## 2. Переменные на VPS

В `.env` на сервере (для сборки miniapp в Docker):

```env
VITE_WALLETCONNECT_PROJECT_ID=ваш_project_id
VITE_APP_URL=https://clutch-duel.ru
VITE_SOLANA_NETWORK=devnet
VITE_SOLANA_RPC_URL=https://api.devnet.solana.com
```

`Dockerfile.nginx` собирает miniapp — переменные должны быть в `.env` **до** `docker compose build nginx`.

## 3. Деплой

```bash
cd ~/clutch
git pull
docker compose -f docker-compose.yml -f docker-compose.ssl.yml build nginx
docker compose -f docker-compose.yml -f docker-compose.ssl.yml up -d nginx
```

## 4. Flow (как gmgn)

1. Экран с **иконками кошельков**: Phantom, Trust, MetaMask, QR/Другие.
2. Тап по Phantom → deep link в приложение кошелька (WalletConnect v2).
3. Подтверждение подключения → подпись SIWS → `POST /auth/wallet/link`.
4. Для on-chain tx снова подключи кошелёк (баннер в комнате дуэли).

Если «ничего не происходит» — проверь красную плашку «Project ID не в сборке» и пересобери nginx.

## 5. Allowed origins (Reown)

В настройках проекта добавь:

- `https://clutch-duel.ru`
- `https://www.clutch-duel.ru` (если используешь)

Без этого WC может не открываться в Telegram WebView.
