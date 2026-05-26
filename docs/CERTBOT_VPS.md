# Certbot на VPS — проверка и перевыпуск

CLUTCH: nginx в Docker. Certbot обычно ставят **на хост** (Ubuntu), сертификаты лежат в `/etc/letsencrypt/`.  
Если certbot прошёл, а в браузере всё ещё «Небезопасно» — чаще всего **nginx в Docker не слушает 443** и не подключён к сертификату.

Замени `YOUR_DOMAIN.ru` на свой домен везде ниже.

---

## 1. Быстрая диагностика (скопируй блоком)

```bash
DOMAIN="YOUR_DOMAIN.ru"
cd ~/clutch   # путь к проекту

# Сертификат на диске?
sudo ls -la /etc/letsencrypt/live/$DOMAIN/ 2>/dev/null || echo "НЕТ папки сертификата"

# Срок действия
sudo openssl x509 -in /etc/letsencrypt/live/$DOMAIN/fullchain.pem -noout -dates 2>/dev/null \
  || echo "fullchain.pem не найден"

# Список всех certbot-сертификатов
sudo certbot certificates

# Кто слушает 80 и 443 на хосте
sudo ss -tlnp | grep -E ':80 |:443 '

# Ответ по HTTPS снаружи (с VPS)
curl -sI "https://$DOMAIN/" | head -8
curl -sI "http://$DOMAIN/" | head -5

# Проверка TLS (даты, кому выдан)
echo | openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" 2>/dev/null \
  | openssl x509 -noout -subject -issuer -dates

# Docker CLUTCH
docker compose ps
docker compose port nginx 80 2>/dev/null
docker compose port nginx 443 2>/dev/null || echo "nginx:443 не проброшен"
```

### Как читать результат

| Проверка | Ок | Проблема |
|----------|-----|----------|
| `live/$DOMAIN/fullchain.pem` | файлы есть | certbot не выпускал или другой домен |
| `notAfter` в будущем | сертификат живой | истёк → renew |
| `curl https://` → `HTTP/2 200` или `301` | HTTPS работает | нет 443 или неверный nginx |
| `curl http://` без редиректа на https | — | нет redirect 80→443 |
| `docker compose port nginx 443` | `0.0.0.0:443` | **443 не проброшен в compose** — типичная причина |
| `ss -tlnp :443` | `docker-proxy` или `nginx` | никто не слушает 443 |

---

## 2. Проверить, что certbot установлен

```bash
certbot --version
# или
sudo certbot --version
```

Установка (если нет):

```bash
sudo apt update
sudo apt install -y certbot
```

---

## 3. Посмотреть все сертификаты и сроки

```bash
sudo certbot certificates
```

Пример хорошего вывода:

```text
Certificate Name: YOUR_DOMAIN.ru
  Domains: YOUR_DOMAIN.ru www.YOUR_DOMAIN.ru
  Expiry Date: 2026-06-01 (VALID: 89 days)
  Certificate Path: /etc/letsencrypt/live/YOUR_DOMAIN.ru/fullchain.pem
```

Если `INVALID` / `EXPIRED` — нужен renew или новый выпуск (раздел 5).

Детали одного сертификата:

```bash
DOMAIN="YOUR_DOMAIN.ru"
sudo openssl x509 -in /etc/letsencrypt/live/$DOMAIN/fullchain.pem -noout -text \
  | grep -E 'Subject:|Issuer:|Not Before|Not After'
```

---

## 4. Проверить HTTPS «как Telegram»

```bash
DOMAIN="YOUR_DOMAIN.ru"

# Заголовки
curl -vI "https://$DOMAIN/" 2>&1 | head -30

# Mini App (должен быть JS-бандл)
curl -s "https://$DOMAIN/" | grep -oE '/assets/[^"]+\.js' | head -1

# API health через nginx
curl -s "https://$DOMAIN/health"

# Ошибки сертификата (если self-signed / просрочен)
curl -vI "https://$DOMAIN/" 2>&1 | grep -i 'ssl\|certificate\|expire'
```

В браузере на телефоне адрес должен начинаться с **`https://`** и быть **без** «Небезопасно».

---

## 5. Обновить (renew) существующий сертификат

Сначала тест без изменений:

```bash
sudo certbot renew --dry-run
```

Если `dry-run` успешен:

```bash
sudo certbot renew
```

Принудительно перевыпустить (если dry-run ок, а браузер ругается):

```bash
sudo certbot renew --force-renewal
```

После renew **перезагрузи nginx**, который использует сертификат:

```bash
# если nginx на хосте
sudo systemctl reload nginx

# если CLUTCH в Docker (после настройки 443, см. раздел 7)
cd ~/clutch && docker compose restart nginx
```

Логи certbot:

```bash
sudo tail -50 /var/log/letsencrypt/letsencrypt.log
```

---

## 6. Выпустить сертификат заново (если не было / протух / сменился домен)

### Вариант A: standalone (проще всего)

**Порт 80 должен быть свободен.** Останови всё, что его занимает:

```bash
cd ~/clutch
docker compose stop nginx
# если на хосте ещё nginx/apache:
# sudo systemctl stop nginx
```

Выпуск:

```bash
DOMAIN="YOUR_DOMAIN.ru"
sudo certbot certonly --standalone \
  -d "$DOMAIN" \
  -d "www.$DOMAIN" \
  --agree-tos \
  -m your@email.com \
  --non-interactive
```

Интерактивно (если нужны вопросы certbot):

```bash
sudo certbot certonly --standalone -d YOUR_DOMAIN.ru -d www.YOUR_DOMAIN.ru
```

Проверка:

```bash
sudo ls -la /etc/letsencrypt/live/YOUR_DOMAIN.ru/
sudo certbot certificates
```

Запусти CLUTCH снова:

```bash
cd ~/clutch
docker compose up -d nginx
```

### Вариант B: webroot (nginx уже отдаёт файлы на 80)

```bash
sudo mkdir -p /var/www/certbot
# в nginx на 80 должен быть:
#   location /.well-known/acme-challenge/ { root /var/www/certbot; }

sudo certbot certonly --webroot -w /var/www/certbot \
  -d YOUR_DOMAIN.ru -d www.YOUR_DOMAIN.ru
```

### Вариант C: только один домен (без www)

```bash
sudo certbot certonly --standalone -d YOUR_DOMAIN.ru
```

### Ошибки при выпуске

```bash
# DNS указывает на этот VPS?
dig +short YOUR_DOMAIN.ru
curl -s ifconfig.me   # IP сервера

# порт 80 слушает кто-то другой?
sudo ss -tlnp | grep ':80 '

# лог
sudo tail -100 /var/log/letsencrypt/letsencrypt.log
```

Частые причины: DNS ещё не обновился, порт 80 занят, firewall закрыл 80/443:

```bash
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

---

## 7. Подключить сертификат к nginx в Docker (важно для CLUTCH)

Сейчас в `docker-compose.yml` проброшен только **80**. Certbot на хосте сам по себе HTTPS не включит.

### 7.1. Проброс 443 и монтирование сертификатов

В `docker-compose.yml` у сервиса `nginx`:

```yaml
    ports:
      - "${HTTP_PORT:-80}:80"
      - "${HTTPS_PORT:-443}:443"
    volumes:
      - /etc/letsencrypt:/etc/letsencrypt:ro
      - ./deploy/nginx/nginx-ssl.conf:/etc/nginx/nginx.conf:ro
```

(файл `deploy/nginx/nginx-ssl.conf` — из примера в репо, с **твоим доменом** в `server_name` и путях к pem).

### 7.2. Создать nginx-ssl.conf на VPS

```bash
cd ~/clutch
DOMAIN="YOUR_DOMAIN.ru"
sed "s/your-domain.ru/$DOMAIN/g" deploy/nginx/nginx-ssl.example.conf \
  | sudo tee deploy/nginx/nginx-ssl.conf > /dev/null
nano deploy/nginx/nginx-ssl.conf   # проверь server_name и пути
```

### 7.3. Перезапуск

```bash
docker compose up -d --build nginx
docker compose port nginx 443
curl -sI "https://YOUR_DOMAIN.ru/" | head -5
```

### 7.4. Авто-renew и reload nginx

```bash
sudo crontab -e
```

Добавь (путь к clutch поправь):

```cron
0 3 * * * certbot renew --quiet --deploy-hook "cd /home/ubuntu/clutch && docker compose restart nginx"
```

Или systemd timer certbot (обычно уже есть после `apt install certbot`):

```bash
sudo systemctl status certbot.timer
sudo systemctl list-timers | grep certbot
```

---

## 8. .env и BotFather после HTTPS

```env
PUBLIC_URL=https://YOUR_DOMAIN.ru
API_PUBLIC_URL=https://YOUR_DOMAIN.ru
MINIAPP_PUBLIC_URL=https://YOUR_DOMAIN.ru
CORS_ORIGINS=https://YOUR_DOMAIN.ru
TELEGRAM_WEBHOOK_PUBLIC_URL=https://YOUR_DOMAIN.ru/telegram/webhook
```

BotFather:

- `/setdomain` → `YOUR_DOMAIN.ru`
- `/setmenubutton` → `https://YOUR_DOMAIN.ru`

```bash
cd ~/clutch && make redeploy-backend
```

---

## 9. Чеклист «всё ок»

```bash
DOMAIN="YOUR_DOMAIN.ru"
sudo certbot certificates | grep -A2 "$DOMAIN"
curl -sI "https://$DOMAIN/" | head -3
curl -s "https://$DOMAIN/health"
curl -s "https://$DOMAIN/" | grep assets
docker compose port nginx 443
```

В Telegram: закрыть Mini App полностью → открыть снова из бота.

---

## 10. Если certbot на хосте, а CLUTCH в Docker без 443

Типичная картина: **certbot успешен**, в браузере по IP/домену **http** и «Небезопасно», в Telegram **белый экран**.

Решение: раздел **7** (443 + `nginx-ssl.conf` + volume `/etc/letsencrypt`).

Альтернатива: Cloudflare proxy (HTTPS на краю, origin :80) — [HTTPS_TELEGRAM.md](./HTTPS_TELEGRAM.md).
