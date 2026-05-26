# flow-service (Go)

Backend Go do app Flow. Port idiomático do Quarkus/Java service.

- **HTTP:** `net/http` + `chi`
- **Persistência:** AWS SDK v2 / DynamoDB Single-Table (mesma tabela do projeto Java)
- **Deploy:** Lambda + API Gateway (HTTP API v2). Mesmo router roda local como servidor HTTP.

## Local stack (Docker)

```bash
docker compose up
```

| Serviço | URL | Descrição |
|---------|-----|-----------|
| Backend  | http://localhost:8080 | Go com hot reload via `air` |
| Health   | http://localhost:8080/q/health | `{"status":"UP"}` |
| DynamoDB | http://localhost:8000 | DynamoDB Local (in-memory) |

A tabela `flow-table` (PK/SK + GSI1) é criada automaticamente na primeira subida.

## Sem Docker

```bash
make run               # server local em :8080
make test              # testes unitários
make build             # binário em bin/server
make lambda            # bootstrap zipado para Lambda (arm64)
```

## Layout

```
cmd/
├── server/        # entry point HTTP local
└── lambda/        # entry point Lambda (mesmo router via chiadapter)
internal/
├── api/           # handlers HTTP (router + middlewares)
├── command/       # write paths (CreateAccount, PostTransaction, …)
├── query/         # read paths (AccountQuery, TransactionQuery, …)
├── service/       # cross-cutting (SystemAccounts, Dashboard, Report)
├── dto/           # request/response structs
├── dynamodb/      # client + key helpers (single-table)
└── config/        # env loading
```

## Status da tradução

| Domínio | Status |
|---------|--------|
| Infra (router, config, DynamoDB client, keys) | ✅ |
| Auth (signup/login/me) | ⏳ |
| Ledger (accounts + transactions) | ⏳ |
| Categories | ⏳ |
| Dashboard | ⏳ |
| Reports | ⏳ |
| Planning (budgets/goals/params/salary) | ⏳ |
