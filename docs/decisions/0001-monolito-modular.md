# ADR-0001: Monólito modular

- **Status:** aceito
- **Data:** 2026-11-11

## Contexto

O e-commerce da ótica começa com um time de uma pessoa, tráfego baixo e
domínios ainda em descoberta (catálogo, estoque, pedidos, pagamentos, entrega).
Queremos baixo custo operacional, mas sem impedir a extração de serviços quando
houver necessidade real.

## Decisão

Um único binário Go (`cmd/api`), organizado por domínios em `internal/`. Cada
domínio segue camadas próprias:

```text
internal/<dominio>/
├── domain/          # entidades e regras de negócio, sem dependências externas
├── application/     # casos de uso; orquestra o domínio
├── infrastructure/  # banco, storage, clientes externos
└── transport/http/  # handlers HTTP
```

Regras:

- Um domínio não acessa tabelas nem pacotes internos de outro. A comunicação
  acontece por interfaces expostas pela camada `application`.
- Pastas de domínio são criadas quando o domínio começa a ser implementado.

## Alternativas consideradas

- **Microsserviços desde o início:** isolamento forte, porém custo alto de
  infraestrutura, deploy, observabilidade e consistência de dados. Prematuro.
- **Monólito sem fronteiras:** mais rápido no começo, mas dificulta a extração
  futura e mistura regras de domínios diferentes.

## Consequências

- Um deploy, um banco, operação simples.
- As fronteiras dependem de disciplina (e de revisão de código). Um linter de
  dependências entre pacotes pode ser adicionado se elas começarem a vazar.
- Extrair um domínio no futuro significa trocar chamadas internas por HTTP ou
  mensagens, sem reescrever as regras de negócio.
