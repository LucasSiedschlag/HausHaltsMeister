# API — Convencoes Gerais

Este documento centraliza as regras comuns a todos os endpoints.

## Escopo e autenticacao

- Toda request autenticada deve fornecer `user_id` valido (token/sessao).
- Endpoints sensiveis exigem role `owner` ou `editor`.
- Endpoints de leitura podem ser acessados por `viewer`.

## Ledger first

- Quase todos os recursos sao escopados por `ledger_id`.
- Nunca buscar recursos somente por `id` sem validar o ledger.

## Datas e meses

- Datas: `YYYY-MM-DD`.
- Mes de referencia: `YYYY-MM-01` (primeiro dia do mes).

## Valores

- Valores monetarios em centavos: `amount_cents` (inteiro).
- `entries.amount_cents` sempre positivo; IN/OUT vem de `categories.direction`.

## Valores enumerados

- Todos os valores enumerados sao `varchar` com validacao via `CHECK` no banco.
- Valores validos devem ser tratados no service layer e rejeitados com `422`.

## Metodos e atualizacoes

- `POST` cria.
- `GET` lista ou detalha.
- `PATCH` atualiza parcialmente.
- `DELETE` remove (preferir soft delete quando aplicavel).

## Codigos de resposta (padrao)

- `400` entrada invalida
- `401` nao autenticado
- `403` sem permissao
- `404` nao encontrado (ou recurso de outro ledger)
- `409` conflito/duplicidade
- `422` validacao de dominio

## Atualizacao de timestamps

- `created_at` sempre definido no banco.
- `updated_at` deve ser setado pela aplicacao em toda atualizacao.
