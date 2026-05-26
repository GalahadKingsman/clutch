# HTTPS для Telegram Mini App (обязательно)

## Симптомы

| Где открываешь | Что видишь |
|----------------|------------|
| Браузер по `http://домен` | Тёмный экран: «Открой через Telegram» — **это норма** |
| Telegram Mini App | Белый экран, квадратики загрузки — **нет HTTPS** |

Telegram **не запускает** Mini App по обычному HTTP. В браузере сайт может открываться, в WebView Telegram — нет.

На скриншоте браузера: **«Небезопасно»** → домен без SSL. Это и есть причина.

---

## Быстрый путь: Cloudflare (рекомендуется)

1. Добавь домен в [Cloudflare](https://dash.cloudflare.com).
2. DNS: `A` запись → IP твоего VPS, **прокси включён** (оранжевое облако).
3. SSL/TLS → режим **Flexible** (если на VPS только порт 80) или **Full** (если поставишь сертификат на nginx).
4. В `.env` на VPS **только `https://`**:

```env
PUBLIC_URL=https://твой-домен.ru
API_PUBLIC_URL=https://твой-домен.ru
MINIAPP_PUBLIC_URL=https://твой-домен.ru
CORS_ORIGINS=https://твой-домен.ru
TELEGRAM_WEBHOOK_PUBLIC_URL=https://твой-домен.ru/telegram/webhook
```

5. **BotFather** (@BotFather):
   - `/setmenubutton` → URL: `https://твой-домен.ru` (тот же, что `MINIAPP_PUBLIC_URL`)
   - Без `http://`, без лишнего слэша в конце

6. Перезапуск:

```bash
docker compose --env-file .env up -d --build nginx api bot
```

7. Проверка:

```bash
curl -sI https://твой-домен.ru/ | head -5
# HTTP/2 200, не редирект на http
curl -s https://твой-домен.ru/ | grep -o '/assets/[^"]*\.js' | head -1
```

8. В Telegram: полностью закрой Mini App и открой снова из бота (кэш WebView).

---

## Certbot на VPS (уже ставили)

Полный чеклист команд: **[CERTBOT_VPS.md](./CERTBOT_VPS.md)** — проверка сертификата, renew, перевыпуск, подключение к Docker.

Кратко после выпуска cert на хосте:

```bash
# проверка
sudo certbot certificates
curl -sI https://твой-домен.ru/ | head -5

# CLUTCH: проброс 443 + ssl-конфиг
cp deploy/nginx/nginx-ssl.example.conf deploy/nginx/nginx-ssl.conf
# отредактируй домен в nginx-ssl.conf
docker compose -f docker-compose.yml -f docker-compose.ssl.yml up -d --build nginx
```

---

## Чеклист

- [ ] В браузере адрес начинается с `https://` и замок (не «Небезопасно»)
- [ ] BotFather Web App URL = `MINIAPP_PUBLIC_URL` (один в один)
- [ ] `curl https://домен/health` отвечает
- [ ] После деплоя: `make redeploy-backend`
