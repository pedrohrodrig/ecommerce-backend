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

## Decisões de arquitetura

Decisões relevantes viram um ADR em `docs/decisions/`, a partir do
[template](docs/decisions/0000-template.md).
