import { Routes } from '@angular/router';
import { loadRemoteModule } from '@angular-architects/native-federation';

import { ROUTES } from './core/constants/app.constants';

export const routes: Routes = [
  {
    path: ROUTES.DASHBOARD,
    loadComponent: () =>
      loadRemoteModule('dashboard', './Component').then((m) => m.App),
  },
  {
    path: ROUTES.TRANSACTIONS,
    loadComponent: () =>
      loadRemoteModule('transactions', './Component').then((m) => m.App),
  },
  {
    path: ROUTES.ACCOUNTS,
    loadComponent: () =>
      loadRemoteModule('accounts', './Component').then((m) => m.App),
  },
  {
    path: ROUTES.REPORTS,
    loadComponent: () =>
      loadRemoteModule('reports', './Component').then((m) => m.App),
  },
];
