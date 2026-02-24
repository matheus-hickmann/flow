# Plan Service

Serviço Plan com **Java 25**, **Spring Boot 3.5** e suporte a **GraalVM Native Image**.

## Requisitos

- **JDK 25** (GraalVM 25 recomendado para build nativo)
- **Maven 3.6.3+**

### GraalVM para Native Image

Para gerar o executável nativo, use [GraalVM 25](https://www.graalvm.org/) ou JDK 25 com `native-image` instalado:

```bash
gu install native-image
```

## Executar na JVM

```bash
./mvnw spring-boot:run
```

Ou com Maven instalado:

```bash
mvn spring-boot:run
```

A aplicação sobe em `http://localhost:8081`.

## Build Native (GraalVM)

Gera um executável nativo (startup rápido, menor uso de memória):

```bash
./mvnw -Pnative native:compile
```

O binário será gerado em `target/plan-service`.

Executar:

```bash
./target/plan-service
```

### Build com Docker (Native Image)

Usando Cloud Native Buildpacks (não precisa do GraalVM instalado localmente):

```bash
./mvnw -Pnative spring-boot:build-image
docker run -p 8081:8081 plan-service:1.0.0-SNAPSHOT
```

## Endpoints

### Health
| Endpoint            | Descrição           |
|---------------------|---------------------|
| `GET /api/health`   | Health check do serviço |
| `GET /actuator/health` | Actuator health |
| `GET /actuator/info`    | Actuator info  |

### Budgeting (limites por categoria)
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST   | `/api/v1/budgets`     | Cadastrar limite (categoria, limitType: ABSOLUTE \| PERCENTAGE, limitValue) |
| GET    | `/api/v1/budgets`     | Listar todos |
| GET    | `/api/v1/budgets/{id}`| Buscar por id |
| PUT    | `/api/v1/budgets/{id}`| Atualizar limite |

### Metas de investimento
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST   | `/api/v1/investment-goals`     | Cadastrar meta (name, expectedReturnRate, monthlyContribution, targetAmount opcional) |
| GET    | `/api/v1/investment-goals`     | Listar todas |
| GET    | `/api/v1/investment-goals/{id}`| Buscar por id |
| PUT    | `/api/v1/investment-goals/{id}`| Atualizar meta |

### Parâmetros econômicos (Selic / IPCA)
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET    | `/api/v1/economic-parameters` | Obter expectativas atuais (404 se ainda não cadastrado) |
| PUT    | `/api/v1/economic-parameters` | Criar ou atualizar (selicRate, ipcaRate) |

## Estrutura

```
src/main/java/com/flow/plan/
├── PlanServiceApplication.java
├── controller/
│   ├── HealthController.java
│   ├── BudgetLimitController.java
│   ├── InvestmentGoalController.java
│   └── EconomicParametersController.java
├── dto/
├── exception/
├── model/entity/
├── repository/
└── service/
```

## Perfis

- **default** – execução normal na JVM
- **native** – usado no build GraalVM Native Image; ativa `application-native.yml` (ex.: JMX desabilitado)
