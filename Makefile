.PHONY: dev-api dev-web web build test generate docker

dev-api:
	set -a; [ -f .env ] && . ./.env; set +a; \
	APP_PASSWORD=$${APP_PASSWORD:-dev} APP_COOKIE_SECURE=false go run ./cmd/gonteractive

dev-web:
	cd web && pnpm dev

web:
	cd web && pnpm install --frozen-lockfile && pnpm build && touch dist/.gitkeep

build: web
	CGO_ENABLED=0 go build -trimpath -o bin/gonteractive ./cmd/gonteractive

test:
	go test -race ./...
	cd web && pnpm lint && pnpm exec tsc -b

generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate

docker:
	docker compose up -d --build
