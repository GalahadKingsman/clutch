# Phase 2 — Dispute & AI

## Цель

Спор после claim, загрузка пруфов, ИИ-судья, экран вердикта, 5-минутное окно апелляции, уведомления в боте.

On-chain `settle_ai` **отложен** — settlement в БД (как mutual confirm в Phase 1).

## Миграция

```bash
# на VPS после pull
docker compose run --rm migrate up
```

Файл: `migrations/000003_phase2.up.sql`

## API (JWT + wallet)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/duels/{id}/dispute` | Открыть спор (не claimer) |
| GET/POST | `/duels/{id}/proofs` | Список / загрузка (multipart `file`) |
| POST | `/duels/{id}/judge` | Запуск ИИ-судьи |
| GET | `/duels/{id}/verdict` | Вердикт (+ авто-finalize после окна) |
| POST | `/duels/{id}/verdict/finalize` | Досрочно зафиксировать (победитель) |
| POST | `/duels/{id}/appeal` | Апелляция → `human_arbitration` |
| POST | `/ai/clarify-condition` | Уточнение условия при создании |
| GET | `/files/*` | Публичная раздача пруфов |

## Env

```env
UPLOAD_DIR=data/uploads
OPENAI_API_KEY=          # опционально
OPENAI_MODEL=gpt-4o-mini
TELEGRAM_ARBITER_CHAT_ID=0
```

Docker: volume `uploads_data` → `/app/data/uploads`.

## Mini App

- `/duel/:id` — кнопки «Оспорить», переход в арбитраж
- `/duel/:id/arbitration` — пруфы + «Запустить ИИ-судью»
- `/duel/:id/verdict` — вердикт, апелляция, таймер 5 мин
- Create — «Уточнить условие (AI)»

## Деплой

```bash
git pull
# .env: TELEGRAM_ARBITER_CHAT_ID, OPENAI_API_KEY (по желанию)
make redeploy-backend redeploy-miniapp
```

## Flow

1. `POST /claim` → `awaiting_claim`
2. Соперник: `confirm` → `mutual_settled` **или** `dispute` → `disputed`
3. Пруфы → `judge` → `appeal_window` (5 мин)
4. Проигравший: `appeal` → Phase 3 (human) **или** окно истекло → `settled`

## Чеклист

- [x] Migration 000003
- [x] Dispute handlers + storage
- [x] AI clarify/judge (OpenAI + fallback)
- [x] Mini App screens
- [x] Bot notifications (verdict, dispute, settle)
- [ ] On-chain `settle_ai` (Phase 2.5 / Sprint)
