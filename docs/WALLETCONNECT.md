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

## 6. Известные проблемы кошельков

### Trust: «Some of the required chains are not supported yet»

На **Solana devnet** Trust Wallet через WalletConnect часто **не поддерживает** сеть. Для тестов используй **Phantom** (иконка первая).

На mainnet (`VITE_SOLANA_NETWORK=mainnet-beta`) Trust обычно работает.

### MetaMask / другие: подключилось, но экран привязки не уходит

Исправлено в miniapp: адрес берётся из namespace `solana`, после WC ожидается `walletProvider` и автоматически запрашивается SIWS.

Если кошелёк вернулся в Mini App без подписи — нажми **«Продолжить привязку»**.

MetaMask должен иметь включённый **Solana** в настройках; иначе адрес Solana не придёт.

### Telegram: бесконечная загрузка «Continue in MetaMask»

В Mini App внутри Telegram `window.open(metamask://…)` не открывает кошелёк корректно — сессия WC зависает.

**Решение в CLUTCH:** перехват ссылок через `Telegram.WebApp.openLink`, на мобильном Telegram показываются только **Phantom** и **QR**. MetaMask/Trust — в обычном браузере.

Если всё же зависло: «Отменить подключение» → полностью закрыть MetaMask → снова **Phantom**.

### Phantom: «Not Detected» (кошелёк установлен)

В Telegram WebView Phantom **не определяется** как расширение — Reown ошибочно ведёт в App Store. CLUTCH обходит это: при нажатии Phantom создаётся WC-сессия и открывается **установленный** Phantom через deep link (не App Store). Если не сработало — **«Открыть установленный Phantom»** или QR в модалке.
