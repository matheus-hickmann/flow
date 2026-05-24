import { Routes } from '@angular/router';

import { ROUTES } from './core/constants/app.constants';
import { authGuard } from './core/guards/auth.guard';
import { LoginComponent } from './pages/auth/login.component';
import { SignupComponent } from './pages/auth/signup.component';

export const routes: Routes = [
  { path: 'login', component: LoginComponent },
  { path: 'criar-conta', component: SignupComponent },
  {
    path: ROUTES.DASHBOARD,
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/dashboard/dashboard.component').then((m) => m.DashboardComponent),
  },
  {
    path: ROUTES.TRANSACTIONS,
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/transactions/transactions.component').then((m) => m.TransactionsComponent),
  },
  {
    path: ROUTES.ACCOUNTS,
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/accounts/accounts.component').then((m) => m.AccountsComponent),
  },
  {
    path: ROUTES.REPORTS,
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/reports/reports.component').then((m) => m.ReportsComponent),
  },
  {
    path: ROUTES.CATEGORIES,
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/categories/categories.component').then((m) => m.CategoriesComponent),
  },
];
