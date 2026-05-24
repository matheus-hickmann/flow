# CLAUDE.md - Full Stack Guide (Go & Angular)

## Role & Standards
Expert Full Stack Developer focused on Go and Angular (21+).
Adheres to SOLID principles, Clean Code, and High-Performance patterns.

---

## Project Layout

| Diretório | Descrição |
|-----------|-----------|
| `flow-app/` | SPA Angular 21 (standalone components, signals, Tailwind) |
| `../flow-service-go/` | API Go (chi, DynamoDB Single-Table) |

Stack: Angular 21, Go 1.23, AWS DynamoDB, Docker, Lambda + API Gateway.

---

## GLOBAL STANDARDS
- **Indentation**: 2 spaces (TS/HTML); tabs (Go — `gofmt`).
- **Quotes**: Single quotes (`'`) for TS/Angular; double quotes (`"`) for Go strings.
- **Naming Conventions**:
    - TS/Angular files: `kebab-case` (e.g., `account-list.component.ts`).
    - Go files: `snake_case` (e.g., `account_service.go`).
    - Classes/Interfaces/Structs: `PascalCase`.
    - Methods/Variables: `camelCase` (TS) / `camelCase` exported, `camelCase` unexported (Go).
    - Constants: `ALL_CAPS_WITH_UNDERSCORES` (TS) / `PascalCase` (Go exported).

---

## BACKEND: Go 1.23
- **HTTP**: `net/http` + `github.com/go-chi/chi/v5`.
- **Architecture**: `internal/api` (handlers) -> `internal/service` (business logic) -> `internal/dynamodb` (persistence).
- **DTOs**: plain structs in `internal/dto/`.
- **Persistence**: AWS SDK v2 / DynamoDB Single-Table. Keys helpers in `internal/dynamodb/keys.go`.
- **Auth**: JWT middleware in `internal/api/middleware/auth.go`. Dev mode: `Bearer dev-<userId>`.
- **Testing**:
    - Standard `testing` package + table-driven tests.
    - Pattern: `TestServiceMethod_scenario`.
    - Fake DynamoDB in `internal/service/fake_dynamo_test.go`.
- **Build**:
    - Run: `make run` (`:8080`)
    - Test: `make test`
    - Build: `make build`
    - Lambda zip: `make lambda`

---

## FRONTEND: Angular 21
- **Reactivity**: Use **Signals** for state management. Avoid manual RxJS subscriptions in components.
- **DI**: Use the `inject()` function instead of constructor parameters.
- **Components**: Use **Standalone Components** exclusively.
- **Templates**: Use `@defer` for lazy loading.
- **Performance**: Use `NgOptimizedImage` and `trackBy` in loops.
- **Styling**: Tailwind CSS for utility-first design; SASS for custom components.
- **Build**:
    - Install: `npm install`
    - Start: `npm start` (`:4200`)
    - Test: `npm test`
    - Build: `npm run build`

---

## LOCAL STACK (Docker)

Sobe tudo a partir de `../flow-service-go/`:

```bash
docker compose up          # DynamoDB Local + init + backend Go em :8080
```

Frontend separado (hot reload):

```bash
cd flow-app && npm start   # :4200
```

---

## TESTING & QUALITY
- **TDD**: Implement TDD for services and business logic.
- **Exclusions**: Do NOT test DTOs or simple getters/setters.
- **Pattern**: Arrange-Act-Assert.
- **Immutability**: Use `readonly` in TypeScript.

---

## ALWAYS AVOID
- Using `any` in TypeScript (strict typing required).
- Business logic in Angular templates or HTTP handlers.
- Direct DOM manipulation in Angular.
- Hardcoded configs (use `environment.ts` on front, env vars on back).
- Global state outside signals (Angular) or context (Go).
