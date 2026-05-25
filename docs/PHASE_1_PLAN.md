# Phase 1 — Core Duel Loop (план)

> **Цель:** полный цикл 1v1 между друзьями: лента → друзья → создать дуэль → вызов → комната → взаимная победа.  
> **Не в Phase 1:** Арена, арбитраж/AI, human appeal, полный indexer mainnet.

---

## Порядок спринтов

| Sprint | Фокус | Exit |
|--------|--------|------|
| **1.1** | DB + модели + Social API | friends, invite, search |
| **1.2** | Duel API (off-chain state) | create, accept, cancel, list, get |
| **1.3** | Feed + bot notifications | лента, пуши на вызов |
| **1.4** | Mini App shell | tab bar, роуты, TMA theme |
| **1.5** | Экраны Feed, Friends, Create, Invite | UI по спекам |
| **1.6** | Комната + WebSocket чат | сообщения, системный «Судья» |
| **1.7** | Mutual win | claim + confirm |
| **1.8** | Price oracle (USD) | Jupiter cache, UI $ |
| **1.9** | On-chain deposit | devnet tx create/accept ✅ (без USDC vault пока) |

---

## Решения по умолчанию (если не ответишь иначе)

| Тема | Решение Phase 1 |
|------|-----------------|
| Эскроу | Сначала **статус в БД** + поля `creator_tx`/`opponent_tx`; on-chain в 1.9 |
| Токен ставки | **USDC devnet** один mint |
| Oracle | Jupiter Price API, кэш Redis 60s |
| Чат | WebSocket `/api/v1/duels/{id}/ws` через nginx upgrade |
| Инвайт | `startapp=invite_{code}` + `POST /friends/accept` |
| Indexer | Polling RPC по `tx_signature` в worker (упрощённый indexer) |

---

## API (добавляем в 1.1–1.7)

Уже есть: auth, wallet gate middleware.

Добавляем:
- `GET /feed/friends`
- `GET|POST /friends/*`
- `GET /users/search`
- `POST|GET /duels/*`
- `WS /duels/{id}/ws`
- `GET /wallet/balances` (stub → RPC позже)

---

## Mini App routes

```
/feed
/friends
/duel/create
/duel/:id          → room
/duel/:id/invite   → incoming challenge view
/wallet
/profile
```

Tab bar: Лента · Друзья · ➕ · Кошелёк · Профиль

---

## Открытые вопросы (см. ответы пользователя)

См. обсуждение в чате.
