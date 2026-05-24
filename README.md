# Flow

Personal finance app: Angular SPA (`flow-app`) + Quarkus backend (`flow-service`) backed by DynamoDB Single-Table.

## Local stack (Docker)

Tudo sobe com um único comando:

```bash
docker compose up
```

| Serviço | URL | Descrição |
|---------|-----|-----------|
| Frontend | http://localhost:4200 | Angular dev server (hot reload) |
| Backend  | http://localhost:8080 | Quarkus dev mode (live reload) |
| OpenAPI  | http://localhost:8080/q/swagger-ui | Swagger UI |
| DynamoDB | http://localhost:8000 | DynamoDB Local (in-memory) |

A inicialização cria a tabela `flow-table` automaticamente (idempotente).

### Comandos úteis

```bash
docker compose up               # sobe tudo em foreground
docker compose up -d            # sobe em background
docker compose down             # desliga tudo
docker compose logs -f backend  # acompanha logs do backend
docker compose restart backend  # reinicia só o backend
```

## Estrutura

| Diretório | Descrição |
|-----------|-----------|
| **flow-app** | SPA Angular 21 (standalone components, signals, Tailwind) |
| **flow-service** | API Quarkus: contas, transações, dashboard, reports, categorias, planning |

## Stack

- **Frontend:** Angular 21, TypeScript, Tailwind CSS, Chart.js
- **Backend:** Quarkus 3.15, Java 21, JAX-RS (blocking), CDI
- **Persistência:** DynamoDB Single-Table (PK/SK + GSI1)
- **Deploy AWS:** Lambda HTTP (API Gateway proxy) + DynamoDB on-demand

## Desenvolvimento sem Docker

**Backend:**
```bash
cd flow-service
mvn quarkus:dev -Dquarkus.profile=local
```

**Frontend:**
```bash
cd flow-app
npm install
npm start
```

Para usar DynamoDB Local sem Docker, suba apenas esse serviço:
```bash
docker compose up -d dynamodb dynamodb-init
```
