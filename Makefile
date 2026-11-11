# Atalhos de desenvolvimento. No Windows, use Git Bash, WSL ou rode os
# comandos equivalentes descritos no README.

BINARY := bin/api
GOLANGCI_LINT_VERSION := v2.15.0

.DEFAULT_GOAL := help

.PHONY: help run build test test-cover lint fmt tidy clean

help: ## Lista os comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

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
