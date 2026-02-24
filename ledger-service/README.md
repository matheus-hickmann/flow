# Ledger Service

Serviço Ledger com **Java 25**, **Spring Boot 3.5** e suporte a **GraalVM Native Image**.

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

A aplicação sobe em `http://localhost:8080`.

## Build Native (GraalVM)

Gera um executável nativo (startup rápido, menor uso de memória):

```bash
./mvnw -Pnative native:compile
```

O binário será gerado em `target/ledger-service`.

Executar:

```bash
./target/ledger-service
```

### Build com Docker (Native Image)

Usando Cloud Native Buildpacks (não precisa do GraalVM instalado localmente):

```bash
./mvnw -Pnative spring-boot:build-image
docker run -p 8080:8080 ledger-service:1.0.0-SNAPSHOT
```

## Endpoints

| Endpoint        | Descrição        |
|----------------|------------------|
| `GET /api/health` | Health check do serviço |
| `GET /actuator/health` | Actuator health |
| `GET /actuator/info`  | Actuator info  |

## Estrutura

```
src/main/java/com/flow/ledger/
├── LedgerServiceApplication.java
└── controller/
    └── HealthController.java
```

## Perfis

- **default** – execução normal na JVM
- **native** – usado no build GraalVM Native Image; ativa `application-native.yml` (ex.: JMX desabilitado)
