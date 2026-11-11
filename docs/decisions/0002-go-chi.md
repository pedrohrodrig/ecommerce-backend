# ADR-0002: Go + Chi

- **Status:** aceito
- **Data:** 2026-11-11

## Contexto

Precisamos de uma linguagem para a API com bom desempenho, deploy simples
(binário estático, imagem Docker pequena) e que sirva de aprendizado.

## Decisão

- **Go** (versão mínima definida em `go.mod`).
- **Chi** como roteador HTTP.

Chi é compatível com `net/http` da biblioteca padrão: handlers e middlewares são
`http.Handler` comuns. Ele adiciona sobre o `http.ServeMux` agrupamento de rotas,
sub-roteadores por domínio e uma coleção de middlewares prontos.

## Alternativas consideradas

- **Apenas `net/http` (ServeMux com padrões do Go 1.22+):** zero dependências,
  mas sem grupos de rotas e middlewares por grupo; teríamos que escrever isso.
- **Gin / Echo / Fiber:** mais recursos, porém com tipos de contexto próprios,
  o que acopla o código ao framework. Fiber nem usa `net/http`.

## Consequências

- Código HTTP portável: trocar o Chi no futuro afeta apenas o roteamento.
- O Chi será adicionado ao `go.mod` na etapa 3 (fundação do backend), quando o
  servidor HTTP for implementado.

## Nota sobre a versão do Go

O `go.mod` declara `go 1.26.0`, não a 1.27. O golangci-lint v2.15 (mais recente
em 2026-11) é compilado com Go 1.26 e se recusa a analisar módulos que exigem
1.27. Subiremos para 1.27 quando o lint suportar.
