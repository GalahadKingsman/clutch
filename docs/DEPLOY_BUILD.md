# Сборка на VPS (Docker) — если «зависает» на npm run build

На этапе `=> [nginx miniapp] RUN npm run build` Vite пишет `rendering chunks...` и кажется, что всё стоит. Чаще всего это **не зависание**, а **нехватка RAM** (OOM) на сервере 1–2 GB.

## Быстрое решение: swap 2 GB

На VPS (один раз):

```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
free -h
```

Потом собирай **по очереди** (не api+nginx вместе — меньше пик RAM):

```bash
cd ~/clutch
git pull

docker compose -f docker-compose.yml -f docker-compose.ssl.yml build api --no-cache
docker compose -f docker-compose.yml -f docker-compose.ssl.yml build nginx --no-cache

docker compose -f docker-compose.yml -f docker-compose.ssl.yml up -d api nginx
```

Сборка nginx miniapp обычно **3–8 минут** на слабом VPS. Секунды могут долго не меняться на `rendering chunks` — это нормально.

## Переменные в .env перед build nginx

```env
VITE_WALLETCONNECT_PROJECT_ID=...
VITE_APP_URL=https://clutch-duel.ru
MINIAPP_PUBLIC_URL=https://clutch-duel.ru
```

## Сборка miniapp на Mac (если VPS всё равно не тянет)

На компьютере:

```bash
cd apps/miniapp
npm install
VITE_WALLETCONNECT_PROJECT_ID=... \
VITE_APP_URL=https://clutch-duel.ru \
VITE_SOLANA_NETWORK=devnet \
VITE_SOLANA_RPC_URL=https://api.devnet.solana.com \
VITE_API_URL=/api/v1 \
npm run build
```

Скопировать `dist` на VPS и собрать nginx без Node-стадии:

```bash
# на VPS в ~/clutch
docker compose -f docker-compose.yml -f docker-compose.ssl.yml -f docker-compose.dist.yml build nginx
docker compose -f docker-compose.yml -f docker-compose.ssl.yml up -d nginx
```

(файл `docker-compose.dist.yml` — только если добавлен в репозиторий)

## Проверка

```bash
docker compose -f docker-compose.yml -f docker-compose.ssl.yml ps
curl -sI https://clutch-duel.ru/ | head -3
```
