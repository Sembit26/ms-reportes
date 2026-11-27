.PHONY: generate migrate-up migrate-status test-api

generate:
	cd apps/api && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

migrate-up:
	cd apps/api && go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir db/migrations postgres "$$DATABASE_URL" up

migrate-status:
	cd apps/api && go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir db/migrations postgres "$$DATABASE_URL" status

test-api:
	cd apps/api && go test ./...
