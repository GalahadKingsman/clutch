.PHONY: help deps go-mod build-api build-bot build-miniapp build migrate-up \
        deploy-backend redeploy-backend deploy-miniapp redeploy-miniapp deploy-all \
        dev-up dev-down logs

help:
	@echo "CLUTCH deploy commands (VPS/production)"
	@echo ""
	@echo "  make deps              - install Go modules + miniapp npm"
	@echo "  make build-miniapp     - Vite production build → apps/miniapp/dist"
	@echo "  make deploy-backend    - docker compose build + up (api, bot, db, nginx)"
	@echo "  make redeploy-backend  - rebuild & restart api + bot + migrate"
	@echo "  make deploy-miniapp    - build miniapp + restart nginx"
	@echo "  make redeploy-miniapp  - same as deploy-miniapp"
	@echo "  make deploy-all        - miniapp + full backend stack"
	@echo "  make migrate-up        - run SQL migrations in docker"
	@echo "  make dev-up            - local infra only (postgres, redis)"

deps: go-mod miniapp-install

go-mod:
	go mod tidy
	go mod download

miniapp-install:
	cd apps/miniapp && npm ci

build-api:
	docker compose build api bot migrate

build-miniapp:
	cd apps/miniapp && npm ci && npm run build

build: build-miniapp build-api

# ─── Production deploy (VPS) ───────────────────────────────
deploy-backend: build-api
	docker compose --env-file .env up -d postgres redis
	docker compose --env-file .env run --rm migrate
	docker compose --env-file .env up -d api bot nginx

redeploy-backend: build-api
	docker compose --env-file .env run --rm migrate
	docker compose --env-file .env up -d --build api bot

deploy-miniapp: build-miniapp
	docker compose --env-file .env up -d nginx

redeploy-miniapp: deploy-miniapp

deploy-all: deploy-miniapp deploy-backend

migrate-up:
	docker compose --env-file .env run --rm migrate

# ─── Local dev ─────────────────────────────────────────────
dev-up:
	docker compose up -d postgres redis

dev-down:
	docker compose down

logs:
	docker compose logs -f api bot nginx
