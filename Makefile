.PHONY: dev backend frontend build test lint fmt proto-gen docker-up docker-down clean

# ─── Full stack ────────────────────────────────────────────────────────────

## Start the Go backend locally
dev: backend

## Run the backend server
backend:
	cd backend && go run ./cmd/server

## Run the frontend dev server
frontend:
	cd frontend && npm run dev

# ─── Build & test ──────────────────────────────────────────────────────────

build:
	$(MAKE) -C backend build
	$(MAKE) -C frontend build

test:
	$(MAKE) -C backend test
	$(MAKE) -C frontend test

lint:
	$(MAKE) -C backend lint
	$(MAKE) -C frontend lint
	$(MAKE) -C proto lint

fmt:
	cd backend && go fmt ./...

# ─── Protos ────────────────────────────────────────────────────────────────

proto-gen:
	$(MAKE) -C proto generate

# ─── Docker ────────────────────────────────────────────────────────────────

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

# ─── Cleanup ───────────────────────────────────────────────────────────────

clean:
	cd backend && go clean -cache
