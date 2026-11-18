BINARY := bin/api
GOLANGCI_LINT_VERSION := v2.15.0
GOOSE_VERSION := v3.30.1
# A URL vai ao Goose pelo ambiente, não pelo argv (visível a outros usuários
# no ps), e é lida pelo shell ($$), para o make não ecoar a senha no terminal.
GOOSE := GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations \
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

.DEFAULT_GOAL := help

.PHONY: help run build test test-cover lint fmt tidy clean migrate-up migrate-down migrate-status require-database-url

help:
	@grep -oE '^[a-z][a-z-]+:' $(MAKEFILE_LIST) | tr -d ':' | grep -v '^require-'

run:
	go run ./cmd/api

build:
	go build -o $(BINARY) ./cmd/api

test:
	go test -race ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

fmt:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) fmt

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out

migrate-up: require-database-url
	$(GOOSE) up

migrate-down: require-database-url
	$(GOOSE) down

migrate-status: require-database-url
	$(GOOSE) status

require-database-url:
	@[ -n "$$(printf '%s' "$$DATABASE_URL" | tr -d '[:space:]')" ] || { echo "DATABASE_URL não definida. Ex.: export DATABASE_URL=postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable" >&2; exit 1; }
