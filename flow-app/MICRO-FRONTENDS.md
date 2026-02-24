# Micro Frontends (Native Federation)

O **flow-app** está dividido em um **shell (host)** e quatro **remotes** (um por tela):

| App         | Projeto        | Porta (dev) | Rota no shell   |
|------------|----------------|-------------|------------------|
| **Shell**  | `flow-app`     | 4200        | —                |
| Dashboard  | `dashboard`    | 4201        | `` (raiz)        |
| Transações | `transactions` | 4202        | `transacoes`     |
| Contas     | `accounts`     | 4203        | `contas`         |
| Relatórios | `reports`      | 4204        | `relatorios`     |

## Desenvolvimento

**Opção 1 – Tudo de uma vez (recomendado):**

```bash
cd flow-app
npm run start:all
```

Isso sobe o shell e os quatro remotes no mesmo terminal (prefixos no log: `d` dashboard, `t` transactions, `a` accounts, `r` reports, `s` shell).

**Opção 2 – Um processo por terminal:**

1. Subir os quatro remotes (um comando por terminal):

   ```bash
   npx ng serve dashboard
   npx ng serve transactions
   npx ng serve accounts
   npx ng serve reports
   ```

2. Subir o shell: `npx ng serve flow-app`

3. Abrir **http://localhost:4200**. O shell carrega cada tela a partir do remote correspondente (manifest em `src/assets/federation.manifest.json`).

## Build

- **Shell:**  
  `ng build flow-app`  
  Saída: `dist/flow-app/`

- **Remotes:**  
  `ng build dashboard`, `ng build transactions`, `ng build accounts`, `ng build reports`  
  Saída: `dist/dashboard/`, `dist/transactions/`, etc.

Em **produção**, ajuste as URLs em `src/assets/federation.manifest.json` (ou troque o arquivo por ambiente) para apontar para as URLs reais de cada remote.

## Tecnologia

- [Native Federation](https://www.npmjs.com/package/@angular-architects/native-federation) (`@angular-architects/native-federation`) com Application Builder (esbuild).
- Shell: `initFederation('assets/federation.manifest.json')` no `main.ts`; rotas usam `loadRemoteModule('nomeRemote', './Component').then(m => m.App)`.
- Cada remote expõe `./Component` (classe `App` em `src/app/app.ts`) e usa `initFederation()` no `main.ts`.
