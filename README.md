# ecommerce-backend

API e regras de negócio do e-commerce da ótica. Monólito modular em Go + Chi.

As decisões de arquitetura (ADRs) ficam no
[Notion do projeto](https://app.notion.com/p/415cc60b2a2980b6b738f80c72276480)
(ADR-0001 monólito modular, ADR-0002 Go + Chi).

> **Fase atual:** fundação do backend. A API lê a configuração do ambiente,
> conecta ao PostgreSQL e sobe o servidor HTTP (Chi) com as rotas em `/v1`, os
> health checks `/healthz` e `/readyz`, log estruturado de cada requisição e
> desligamento gracioso; sem configuração válida ou sem banco acessível ela não
> sobe. Tratamento de erros, middlewares e Docker Compose chegam nas próximas
> tarefas da etapa 3.

## Requisitos

| Ferramenta | Versão | Para quê |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.26 ou superior | compilar e testar |
| [golangci-lint](https://golangci-lint.run/docs/welcome/install/) | v2.15.0 | lint e formatação no editor |
| make (opcional) | qualquer | atalhos; no Windows vem com Git Bash/WSL ou via `winget install GnuWin32.Make` |
| [Docker](https://docs.docker.com/get-started/get-docker/) | qualquer recente | PostgreSQL 18 local e testes de integração |

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
| `make test` | `go test -race ./...` | testes unitários (o `-race` exige CGO; no Windows sem gcc, rode sem ele) |
| `make test-integration` | `go test -race -tags=integration ./tests/integration/...` | testes de integração (exigem Docker) |
| `make test-cover` | `go test -race -tags=integration -coverpkg=./... -coverprofile=coverage.out ./...` e `go tool cover -func=coverage.out` | cobertura de unitário + integração (exige Docker) |
| `make lint` | `golangci-lint run` | lint + verificação de formatação |
| `make fmt` | `golangci-lint fmt` | formata o código |
| `make migrate-up` | `goose up` | aplica as migrations pendentes |
| `make migrate-down` | `goose down` | desfaz a última migration |
| `make migrate-status` | `goose status` | mostra o estado das migrations |

O `make lint`, o `make fmt` e os `make migrate-*` usam `go run` com a versão
fixada, então funcionam mesmo sem o golangci-lint ou o Goose instalados. Sem
make, troque `goose` por `go run github.com/pressly/goose/v3/cmd/goose@v3.30.1`
e defina antes as variáveis `GOOSE_*` (ver [Banco de dados](#banco-de-dados)).
Os `make run` e `make migrate-*` exigem `DATABASE_URL` (ver
[Configuração](#configuração)).

## Configuração

A API lê a configuração de variáveis de ambiente, todas documentadas em
[`.env.example`](.env.example):

| Variável | Obrigatória | Padrão | Para quê |
|---|---|---|---|
| `DATABASE_URL` | sim | | conexão com o PostgreSQL (ver [Banco de dados](#banco-de-dados)) |
| `HTTP_PORT` | não | `8080` | porta HTTP da API |
| `APP_ENV` | não | `development` | ambiente: `development` ou `production` |
| `LOG_LEVEL` | não | `info` | nível mínimo de log: `debug`, `info`, `warn` ou `error` |
| `LOG_PROBES` | não | `true` | loga as requisições a `/healthz` e `/readyz`; use `false` no Kubernetes |

Se alguma variável faltar ou for inválida, a API encerra com código 1 e lista
todos os problemas de uma vez.

Para desenvolver, copie o exemplo e ajuste o que precisar:

```bash
cp .env.example .env
```

O `make` lê o `.env` sozinho (ele não é versionado). Os valores do `.env` têm
prioridade sobre as variáveis do terminal; para sobrescrever um valor numa
execução, passe-o ao make: `make run HTTP_PORT=9090`. Sem make, exporte as
variáveis no terminal antes do `go run`. Em produção não há `.env`: as variáveis
vêm do ambiente.

## Banco de dados

A API usa PostgreSQL 18 (driver [pgx](https://github.com/jackc/pgx)) e lê a
conexão da variável `DATABASE_URL`. Com o banco inacessível, a API encerra com
código 1 e uma mensagem explicando o motivo.

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

Com o `.env` criado (ver [Configuração](#configuração)), aplique as migrations e
rode a API:

```bash
make migrate-up
make run    # loga "server started" com url=http://localhost:8080/v1
curl http://localhost:8080/v1    # {"service":"ecommerce-backend"}
```

Para parar a API, use Ctrl+C (ver [Desligamento](#desligamento)).

Sem make, defina a `DATABASE_URL` no terminal. O Goose lê a conexão das variáveis `GOOSE_*`, que o make já define.
Elas vão pelo ambiente, e não pela linha de comando, para a senha não aparecer
na lista de processos:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable'
export GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations
go run github.com/pressly/goose/v3/cmd/goose@v3.30.1 up
```

No PowerShell (sem make):

```powershell
$env:DATABASE_URL = 'postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable'
$env:GOOSE_DRIVER = 'postgres'; $env:GOOSE_DBSTRING = $env:DATABASE_URL; $env:GOOSE_MIGRATION_DIR = 'migrations'
go run github.com/pressly/goose/v3/cmd/goose@v3.30.1 up
go run ./cmd/api
```

Escreva a `DATABASE_URL` no formato URL e codifique os caracteres especiais do
usuário e da senha (ex.: `@` como `%40`, `/` como `%2F`). A API recusa uma URL
ambígua sem mostrar a senha, mas os erros do Goose podem mostrar a string de
conexão inteira.

Se a porta 5432 já estiver em uso (por exemplo, um PostgreSQL instalado na
máquina), publique outra (`-p 127.0.0.1:5433:5432`) e ajuste a porta na URL.
Para parar e retomar o banco: `docker stop ecommerce-postgres` /
`docker start ecommerce-postgres`.

O `sslmode=disable` serve só para o banco local; fora de localhost, use
`sslmode=verify-full`.

As migrations ficam em [`migrations/`](migrations), em SQL no formato do
[Goose](https://github.com/pressly/goose): um arquivo `NNNNN_descricao.sql` com
as seções `-- +goose Up` e `-- +goose Down`. O `make migrate-down` sem nenhuma
migration aplicada termina com o erro `no current version found` do Goose; é o
esperado.

## Health checks

Ficam fora do `/v1` porque são da infraestrutura (Docker, Kubernetes), não da
API pública:

| Rota | Pergunta que responde | Respostas |
|---|---|---|
| `GET /healthz` | o processo está vivo? Não consulta o banco. | `200 {"status":"ok"}` |
| `GET /readyz` | a API consegue atender? Faz um ping no PostgreSQL (limite de 2s). | `200 {"status":"ok"}` ou `503 {"status":"unavailable"}` |

Quando o `/readyz` falha, o motivo aparece só no log (`readiness check failed`),
nunca na resposta.

## Logs

A API usa o [`log/slog`](https://pkg.go.dev/log/slog) da biblioteca padrão e
escreve na saída padrão. Com `APP_ENV=production` cada linha é um JSON; em
`development` é texto `chave=valor`, mais fácil de ler no terminal. Os campos são
os mesmos nos dois formatos. O nível mínimo vem de `LOG_LEVEL`.

Cada requisição gera uma linha `request` com `method`, `route` (o padrão da rota
no Chi, ex. `/v1/products/{id}`; vazio quando nenhuma rota casa), `path`,
`status`, `duration_ms` e `request_id`. O request id é o do header
`X-Request-Id` quando o cliente manda um válido (até 64 caracteres entre letras,
números, `.`, `_` e `-`); senão a API gera um. Ele sempre volta no header
`X-Request-Id` da resposta, para cruzar uma reclamação com o log.

No Kubernetes, os probes chamam `/healthz` e `/readyz` a cada poucos segundos.
Com `LOG_PROBES=false`, essas requisições deixam de gerar a linha `request`; a
falha do `/readyz` (`readiness check failed`) continua sendo logada.

## Desligamento

Ao receber SIGTERM (o que Docker e Kubernetes enviam) ou Ctrl+C, a API para de
aceitar conexões novas, espera até 20s as requisições em andamento terminarem e
só então fecha o pool do banco e sai com código 0. Se o prazo estourar, ela sai
com código 1.

## Testes

- **Unitários** ficam ao lado do código (`arquivo_test.go`) e não precisam de
  banco nem de Docker: `make test`.
- **De integração** ficam em [`tests/integration/`](tests/integration), atrás da
  build tag `integration`. Eles sobem um `postgres:18-alpine` em container com o
  [Testcontainers](https://golang.testcontainers.org/), aplicam as migrations e
  removem o container no fim: `make test-integration`. Exigem o Docker rodando
  (sem ele, falham dizendo isso); sem a tag, o `go test ./...` nem compila
  esse pacote. Não usam o banco local nem a porta 5432.
- **Cobertura:** `make test-cover` roda unitário + integração juntos, grava
  `coverage.out` e imprime o total. Para ver linha a linha no navegador:
  `go tool cover -html=coverage.out`. No CI, o total aparece no resumo do job
  "Testes de integração" e o `coverage.out` e o `coverage.html` ficam no
  artefato `cobertura` da execução.

No Docker Desktop, se o container `testcontainers/ryuk` cair logo ao subir
(erro `unexpected container status "removing"`), defina
`TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock` antes de rodar os
testes; no Git Bash, prefixe também `MSYS_NO_PATHCONV=1`.

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
cmd/api/                       ponto de entrada da API
internal/platform/config/      leitura e validação das variáveis de ambiente
internal/platform/database/    pool de conexões com o PostgreSQL (pgx)
internal/platform/httpserver/  servidor HTTP, router Chi, health checks, log de requisições e desligamento
internal/platform/logging/     criação do logger (slog) conforme ambiente e nível
migrations/                    migrations SQL (Goose) e embed.go, que as expõe como embed.FS
tests/integration/             testes de integração com PostgreSQL em container
```

As demais pastas (`deployments/`, `tests/e2e/` e os domínios em `internal/`) são
criadas quando recebem o primeiro arquivo. A estrutura completa e o que vai em
cada pasta estão em [CONTRIBUTING.md](CONTRIBUTING.md#estrutura-do-projeto).

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) roda em PRs e pushes para
`dev` e `main`: lint e formatação, `go vet` + testes unitários, testes de
integração com cobertura (ver [Testes](#testes)) e build.

## Contribuindo

Ver [CONTRIBUTING.md](CONTRIBUTING.md) (git flow, Conventional Commits).
