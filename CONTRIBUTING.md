# Contribuindo

## Branches (git flow)

| Branch | Papel |
|---|---|
| `main` | Código em produção. Só recebe merges de `release/*` (e `hotfix/*`). |
| `develop` | Integração. Base de todo trabalho novo. |
| `feature/<descricao>` | Uma funcionalidade ou ajuste. Sai de `develop` e volta por PR. |
| `release/<versao>` | Preparação de uma versão. Sai de `develop`, entra em `main` e volta para `develop`. |

Fluxo comum:

```bash
git checkout develop && git pull
git checkout -b feature/catalog-produtos
# ... commits ...
git push -u origin feature/catalog-produtos
# abrir PR para develop
```

## Commits

Seguimos [Conventional Commits](https://www.conventionalcommits.org/pt-br/):

```text
feat(catalog): adiciona listagem de produtos
fix(http): corrige status de erro de validação
chore: atualiza dependências
docs: adiciona ADR-003
```

## Antes de abrir o PR

```bash
make fmt
make lint
make test
```

O CI roda as mesmas verificações e precisa estar verde para o merge.

## Estrutura do projeto

Pastas são criadas **somente quando recebem o primeiro arquivo** (o Git não
versiona pastas vazias e evitamos arquivos `.gitkeep`). Esta é a estrutura
alvo; ao criar uma pasta, siga o lugar indicado aqui.

```text
ecommerce-backend/
├── cmd/
│   ├── api/                 # binário da API HTTP
│   └── worker/              # binário de jobs assíncronos (quando houver)
├── internal/
│   ├── platform/            # código técnico compartilhado (config, db, logging, http)
│   └── <dominio>/           # um por domínio: identity, catalog, inventory,
│       │                    # customers, cart, orders, payments, shipping
│       ├── domain/          # entidades e regras de negócio, sem dependências externas
│       ├── application/     # casos de uso
│       ├── infrastructure/  # banco, storage, clientes externos
│       └── transport/http/  # handlers HTTP
├── migrations/              # migrations SQL (Goose)
├── deployments/
│   ├── docker/
│   └── kubernetes/
└── tests/
    ├── integration/         # testes com banco/serviços reais
    └── e2e/                 # testes ponta a ponta da API
```

Regras:

- Um domínio não importa pacotes internos de outro domínio nem acessa suas
  tabelas; a comunicação passa pela camada `application`.
- Testes unitários ficam ao lado do código (`arquivo_test.go`); `tests/` é só
  para integração e e2e.
- Mudou a estrutura alvo? Atualize esta seção no mesmo PR.

## Decisões de arquitetura

Decisões relevantes viram um ADR na base
[ADRs do Notion do projeto](https://app.notion.com/p/415cc60b2a2980b6b738f80c72276480).
