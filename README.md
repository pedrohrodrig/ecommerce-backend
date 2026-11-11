# ecommerce-backend

API e regras de negócio do e-commerce da ótica. Monólito modular em Go
([ADR-0001](docs/decisions/0001-monolito-modular.md),
[ADR-0002](docs/decisions/0002-go-chi.md)).

> **Fase atual:** configuração base. O binário apenas inicializa e encerra.
> Servidor HTTP (Chi), PostgreSQL, logging, health checks e Docker chegam na
> etapa 3 (fundação do backend).

## Requisitos

| Ferramenta | Versão | Para quê |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.26 ou superior | compilar e testar |
| [golangci-lint](https://golangci-lint.run/docs/welcome/install/) | v2.15.0 | lint e formatação no editor |
| make (opcional) | qualquer | atalhos; no Windows vem com Git Bash/WSL ou via `winget install GnuWin32.Make` |

Instalar o golangci-lint com Go (funciona no Windows, Linux e macOS):

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
golangci-lint version
```

## Comandos

| make | Sem make | O que faz |
|---|---|---|
| `make run` | `go run ./cmd/api` | executa a API |
| `make build` | `go build -o bin/api ./cmd/api` | compila em `bin/api` |
| `make test` | `go test -race ./...` | testes (o `-race` exige CGO; no Windows sem gcc, rode sem ele) |
| `make lint` | `golangci-lint run` | lint + verificação de formatação |
| `make fmt` | `golangci-lint fmt` | formata o código |

O `make lint` e o `make fmt` usam `go run` com a versão fixada, então funcionam
mesmo sem o golangci-lint instalado.

## Lint e formatação

Tudo é definido em [`.golangci.yml`](.golangci.yml):

- **Formatadores:** `gofmt` + `goimports` (imports do próprio módulo em grupo
  separado). O `golangci-lint run` falha se algum arquivo não estiver formatado;
  `golangci-lint fmt` corrige.
- **Linters:** o conjunto `standard` (errcheck, govet, staticcheck, unused,
  ineffassign) + bodyclose, errorlint, gocritic, gosec, noctx, revive,
  unconvert, unparam.

## VS Code

Ao abrir o repositório, o VS Code sugere as extensões em
[`.vscode/extensions.json`](.vscode/extensions.json):

- **Go** (`golang.go`): autocompletar, testes, debug.
- **EditorConfig** (`editorconfig.editorconfig`): aplica `.editorconfig`.

[`.vscode/settings.json`](.vscode/settings.json) já configura a extensão Go para
usar o `golangci-lint` como linter e como formatador ao salvar, com a mesma
configuração do CI. Requer o golangci-lint instalado no PATH (ver acima).

## Estrutura

```text
cmd/api/            ponto de entrada da API
internal/           domínios (criados sob demanda) e código compartilhado
migrations/         migrations do banco (Goose, etapa 3)
deployments/        docker/ e kubernetes/
tests/              integration/ e e2e/
docs/decisions/     ADRs
```

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) roda em PRs e pushes para
`develop` e `main`: lint e formatação, `go vet` + testes e build.

## Contribuindo

Ver [CONTRIBUTING.md](CONTRIBUTING.md) (git flow, Conventional Commits, ADRs).
