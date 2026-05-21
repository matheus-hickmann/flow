# Flow

Projeto Flow: app Angular (flow-app) + backend serverless (Quarkus, DynamoDB Single-Table, Lambda).

## Executando localmente

1. **Variáveis de ambiente (opcional):** `cp .env.example .env`
2. **DynamoDB Local:** `docker compose up -d`
3. **Criar tabela:** `./scripts/init-dynamodb-local.sh` ou `bash scripts/init-dynamodb-local.sh` (requer AWS CLI)
4. **Ledger-service:** `cd ledger-service && mvn quarkus:dev -Dquarkus.profile=local`
5. **Frontend:** `cd flow-app && npm install && npm start`

- App: **http://localhost:4200**
- API ledger: **http://localhost:8081**
- OpenAPI: **http://localhost:8081/openapi**

Guia completo (pré-requisitos, configuração do ambiente, variáveis): **[docs/LOCAL_SETUP.md](docs/LOCAL_SETUP.md)**.

## Estrutura

| Diretório | Descrição |
|-----------|-----------|
| **flow-app** | SPA Angular (contas, transações, dashboard) |
| **ledger-service** | API Quarkus: escrita atômica + leitura de saldo e extrato (DynamoDB) |
| **account-service** / **planning-service** / **dashboard-service** | Esqueletos Quarkus (mesma tabela) |
| **integration-consumer** | Lambda: consome DynamoDB Stream e notifica (SNS) planning/dashboard |
| **docs** | Documentação (Single-Table, backend serverless, setup local, IntegrationConsumer) |

## Documentação

- [**LOCAL_SETUP.md**](docs/LOCAL_SETUP.md) — Como executar e configurar o ambiente local
- [**SERVERLESS_BACKEND.md**](docs/SERVERLESS_BACKEND.md) — Backend serverless (Quarkus, Lambda, DynamoDB)
- [**DYNAMODB_SINGLE_TABLE.md**](docs/DYNAMODB_SINGLE_TABLE.md) — Design da tabela (PK/SK, GSI, Streams)
- [**INTEGRATION_CONSUMER.md**](docs/INTEGRATION_CONSUMER.md) — Lambda que consome o stream e notifica SNS
