# ecommerce-backend

API e regras de negócio do e-commerce da ótica. Monólito modular em Go + Chi.

As decisões de arquitetura (ADRs) ficam no
[Notion do projeto](https://app.notion.com/p/415cc60b2a2980b6b738f80c72276480)
(ADR-0001 monólito modular, ADR-0002 Go + Chi).

> **Fase atual:** fundação do backend. O binário conecta ao PostgreSQL, confirma
> que o banco responde e encerra; sem banco acessível ele não sobe. Servidor
> HTTP (Chi), pacote de configuração, logging, health checks e Docker Compose
> chegam nas próximas tarefas da etapa 3.

## Requisitos

| Ferramenta | Versão | Para quê |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.26 ou superior | compilar e testar |
| [golangci-lint](https://golangci-lint.run/docs/welcome/install/) | v2.15.0 | lint e formatação no editor |
| make (opcional) | qualquer | atalhos; no Windows vem com Git Bash/WSL ou via `winget install GnuWin32.Make` |
| [Docker](https://docs.docker.com/get-started/get-docker/) | qualquer recente | PostgreSQL 18 local |

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
| `make migrate-up` | `goose up` | aplica as migrations pendentes |
| `make migrate-down` | `goose down` | desfaz a última migration |
| `make migrate-status` | `goose status` | mostra o estado das migrations |

O `make lint`, o `make fmt` e os `make migrate-*` usam `go run` com a versão
fixada, então funcionam mesmo sem o golangci-lint ou o Goose instalados. Sem
make, troque `goose` por `go run github.com/pressly/goose/v3/cmd/goose@v3.30.1`.
Os `make run` e `make migrate-*` exigem `DATABASE_URL` (ver abaixo).

O Goose recebe a conexão pelo ambiente, e não pela linha de comando, para a
senha não aparecer na lista de processos. Sem make, defina antes:

```bash
export GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations
```

```powershell
$env:GOOSE_DRIVER = 'postgres'; $env:GOOSE_DBSTRING = $env:DATABASE_URL; $env:GOOSE_MIGRATION_DIR = 'migrations'
```

## Banco de dados

A API usa PostgreSQL 18 (driver [pgx](https://github.com/jackc/pgx)) e lê a
conexão da variável `DATABASE_URL`. Sem ela, ou com o banco inacessível, a API
encerra com código 1 e uma mensagem explicando o motivo.

Até o Docker Compose chegar, suba o banco local com Docker:

```bash
docker run --name ecommerce-postgres -d \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=ecommerce \
  -p 127.0.0.1:5432:5432 postgres:18-alpine
```

A porta fica só em `127.0.0.1`, fora do alcance da rede. Antes de seguir,
espere `docker logs ecommerce-postgres` mostrar `database system is ready to
accept connections`; na primeira subida a linha aparece duas vezes (a primeira
é do servidor temporário de inicialização), então espere a segunda.

Defina a URL, aplique as migrations e rode a API:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable'
make migrate-up
make run    # imprime "ecommerce-backend: api inicializada"
```

No PowerShell (sem make):

```powershell
$env:DATABASE_URL = 'postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable'
$env:GOOSE_DRIVER = 'postgres'; $env:GOOSE_DBSTRING = $env:DATABASE_URL; $env:GOOSE_MIGRATION_DIR = 'migrations'
go run github.com/pressly/goose/v3/cmd/goose@v3.30.1 up
go run ./cmd/api
```

Se a porta 5432 já estiver em uso (por exemplo, um PostgreSQL instalado na
máquina), publique outra (`-p 127.0.0.1:5433:5432`) e ajuste a porta na URL.
Para parar e retomar o banco: `docker stop ecommerce-postgres` /
`docker start ecommerce-postgres`.

O `sslmode=disable` serve só para o banco local; fora de localhost, use
`sslmode=verify-full`.

As migrations ficam em [`migrations/`](migrations), em SQL no formato do
[Goose](https://github.com/pressly/goose): um arquivo `NNNNN_descricao.sql` com
as seções `-- +goose Up` e `-- +goose Down`.

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
cmd/api/                     ponto de entrada da API
internal/platform/database/  pool de conexões com o PostgreSQL (pgx)
migrations/                  migrations SQL (Goose) e embed.go, que as expõe como embed.FS
```

As demais pastas (`deployments/`, `tests/` e os domínios em `internal/`) são
criadas quando recebem o primeiro arquivo. A estrutura completa e o que vai em
cada pasta estão em [CONTRIBUTING.md](CONTRIBUTING.md#estrutura-do-projeto).

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) roda em PRs e pushes para
`develop` e `main`: lint e formatação, `go vet` + testes e build.

## Contribuindo

Ver [CONTRIBUTING.md](CONTRIBUTING.md) (git flow, Conventional Commits).
