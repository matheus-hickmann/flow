# LedgerApp (flow-app)

SPA do LedgerApp em Angular com Tailwind CSS.

## Pré-requisitos

- Node.js 18+
- npm ou pnpm

## Instalação e execução

```bash
cd flow-app
npm install
npm start
```

Acesse `http://localhost:4200`.

## Estrutura do projeto (Angular best practices)

- **`core/`** – Configuração da aplicação (environment token), constantes, modelos e serviços singleton. Não contém componentes.
- **`shared/`** – Componentes e pipes reutilizáveis (modal, `CurrencyBrlPipe`). Sem lógica de negócio.
- **`layout/`** – Componentes de shell: header (navegação) e FAB menu. Usados pelo `AppComponent`.
- **`features/`** – Funcionalidades por domínio (ex.: `entries` com modais de despesa, receita, planejamento). Export via barrel (`index.ts`).
- **`pages/`** – Containers de rota (dashboard, transações, contas, relatórios). Lazy-loaded nas rotas.

Configuração injetável via token `ENVIRONMENT`; uso de **Signals**, **inject()** e **standalone components**; formatação de moeda centralizada no pipe `currencyBrl`.

## Funcionalidades

- **Dashboard**: resumo geral (Saldo Total, Receitas, Despesas), gastos por categoria (gráfico donut), Orçado vs Realizado e últimos lançamentos.
- **Navegação**: Dashboard, Transações, Contas, Relatórios.
- **FAB**: botão verde no canto inferior direito para adicionar lançamento (Despesa, Receita, Planejamento).

## Upgrade para Angular 21

Para usar Angular 21, atualize as dependências em `package.json` para `^21.0.0` e rode `npm install`.
