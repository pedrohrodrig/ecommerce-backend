# Atalhos de desenvolvimento. No Windows, use Git Bash, WSL ou rode os
# comandos equivalentes descritos no README.

BINARY := bin/api
GOLANGCI_LINT_VERSION := v2.15.0
GOOSE_VERSION := v3.30.1
# A URL é lida pelo shell ($$), para o make não ecoar a senha no terminal.
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir migrations postgres "$$DATABASE_URL"

.DEFAULT_GOAL := help

.PHONY: help run build test test-cover lint fmt tidy clean migrate-up migrate-down migrate-status require-database-url

help: ## Lista os comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

run: ## Executa a API
	go run ./cmd/api

build: ## Compila o binário em bin/api
	go build -o $(BINARY) ./cmd/api

test: ## Roda os testes com detector de race
	go test -race ./...

test-cover: ## Roda os testes e gera relatório de cobertura
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Roda o golangci-lint (versão fixada)
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

fmt: ## Formata o código
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) fmt

tidy: ## Organiza go.mod e go.sum
	go mod tidy

clean: ## Remove artefatos de build
	rm -rf bin coverage.out

migrate-up: require-database-url ## Aplica as migrations pendentes (requer DATABASE_URL)
	$(GOOSE) up

migrate-down: require-database-url ## Desfaz a última migration (requer DATABASE_URL)
	$(GOOSE) down

migrate-status: require-database-url ## Mostra o estado das migrations (requer DATABASE_URL)
	$(GOOSE) status

require-database-url:
	@test -n "$$DATABASE_URL" || { echo "DATABASE_URL não definida. Ex.: export DATABASE_URL=postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable" >&2; exit 1; }
