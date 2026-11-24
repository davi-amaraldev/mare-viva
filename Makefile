.PHONY: setup dev backend frontend migrate seed seed-science seed-demo test build lint

setup:
	cd frontend && npm ci
	cd backend && go mod download

dev:
	$(MAKE) -j2 backend frontend

backend:
	cd backend && set -a; [ ! -f ../.env ] || . ../.env; set +a; go run ./cmd/server

frontend:
	cd frontend && npm run dev -- --host 127.0.0.1

migrate:
	cd backend && set -a; [ ! -f ../.env ] || . ../.env; set +a; go run ./cmd/server migrate

seed: seed-science

seed-science:
	cd backend && set -a; [ ! -f ../.env ] || . ../.env; set +a; go run ./cmd/server seed-science

seed-demo:
	cd backend && set -a; [ ! -f ../.env ] || . ../.env; set +a; go run ./cmd/server seed-demo

test:
	cd backend && go test ./...
	cd frontend && npm run lint && npm run build

build:
	cd backend && go build -o ../mare-viva-api ./cmd/server
	cd frontend && npm run build

lint:
	cd frontend && npm run lint
