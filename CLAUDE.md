# CLAUDE.md - Full Stack Guide (Go & Angular)

## Role & Standards
Expert Full Stack Developer focused on Go and Angular (21+).
Adheres to SOLID principles, Clean Code, and High-Performance patterns.

---

## Project Layout

| Diretório | Descrição |
|-----------|-----------|
| `flow-app/` | SPA Angular 21 (standalone components, signals, Tailwind) |
| `flow-service/` | API Go (chi, DynamoDB Single-Table) |

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

Sobe tudo a partir da raiz do repo:

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

## DESIGN SYSTEM (Angular templates)

Todo template HTML deve usar exclusivamente os tokens Flow. Nunca use classes `neutral-*`, `gray-*`, `blue-*`, `dark:*`, `bg-white`, `bg-primary`, `text-danger` ou similares genéricos.

### Estrutura padrão de página

```html
<div class="font-sans space-y-6">
  <!-- Sub-header -->
  <div class="flex items-center justify-between pb-4 border-b border-flow">
    <div class="flex items-center gap-4">
      <span class="label-eyebrow">EYEBROW</span>
      <h1 class="font-serif text-3xl m-0">Título</h1>
    </div>
    <!-- Ação primária -->
    <button class="h-8 inline-flex items-center gap-1.5 px-3 rounded-lg text-xs font-medium hover:opacity-90"
            style="background: var(--flow-ink); color: var(--flow-bg)">Ação</button>
  </div>
  <!-- Conteúdo em cards -->
</div>
```

### Tokens de cor (via CSS vars, respondem ao dark mode)

| Uso | Classe Tailwind | CSS var |
|-----|-----------------|---------|
| Fundo de card | `bg-flow-paper border border-flow rounded-2xl` | `--flow-paper` |
| Fundo de página | `bg-flow-bg` | `--flow-bg` |
| Texto principal | herdado do `body` | `--flow-ink` |
| Texto secundário | `text-flow-ink-soft` | `--flow-ink-soft` |
| Texto apagado | `text-flow-ink-mute` | `--flow-ink-mute` |
| Borda | `border-flow` | `--flow-line` |
| Borda suave | `border-flow-soft` | `--flow-line-soft` |
| Positivo | `text-flow-pos` | `--flow-pos` |
| Negativo | `text-flow-neg` | `--flow-neg` |
| Pastéis | `bg-flow-mint/peach/butter/lilac/rose/sage/sky` | variáveis correspondentes |

### Componentes recorrentes

**Botão primário:** `h-8 inline-flex items-center px-3 rounded-lg text-xs font-medium hover:opacity-90` + `style="background: var(--flow-ink); color: var(--flow-bg)"`

**Botão secundário/outline:** `h-8 inline-flex items-center px-3 rounded-lg border border-flow text-xs text-flow-ink-soft hover:bg-flow-paper transition`

**Campo de formulário:** `w-full h-9 px-3 rounded-xl border border-flow text-sm focus:outline-none` + `style="background: var(--flow-bg); color: var(--flow-ink)"`

**Label de formulário:** `<label class="label-eyebrow block mb-1">`

**Tab switcher:** `inline-flex border border-flow rounded-lg p-0.5 bg-flow-paper` com botões `[class.bg-flow-ink]="ativo"` / `[class.text-flow-bg]="ativo"`

**Estado vazio:** `bg-flow-paper border border-flow rounded-2xl p-12 text-center` com `<p class="font-mono-num text-[11px] text-flow-ink-mute">`

**Estado de loading:** `bg-flow-paper border border-flow rounded-2xl p-10 text-center` com `<span class="label-eyebrow">`

**Modal customizado (sem app-modal):**
```html
<div class="fixed inset-0 z-50 flex items-center justify-center p-4" style="background: rgba(0,0,0,.5)">
  <div class="w-full max-w-sm rounded-2xl overflow-hidden"
       style="background: var(--flow-paper); border: 1px solid var(--flow-line); box-shadow: var(--flow-shadow-card)">
    <div class="px-6 py-4 border-b border-flow">
      <span class="label-eyebrow">EYEBROW</span>
      <h2 class="font-serif text-2xl mt-0.5">Título</h2>
    </div>
    <div class="px-6 py-5 space-y-4"><!-- conteúdo --></div>
  </div>
</div>
```

**app-modal:** passar sempre `eyebrow="CONTEXTO"` além de `title="Título"`.

---

## ALWAYS AVOID
- Using `any` in TypeScript (strict typing required).
- Business logic in Angular templates or HTTP handlers.
- Direct DOM manipulation in Angular.
- Hardcoded configs (use `environment.ts` on front, env vars on back).
- Global state outside signals (Angular) or context (Go).
- Classes genéricas em templates: `neutral-*`, `gray-*`, `blue-*`, `bg-white`, `dark:*` — use sempre tokens `flow-*`.
