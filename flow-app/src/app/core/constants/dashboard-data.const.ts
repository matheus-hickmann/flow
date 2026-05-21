import type { BudgetVsActualItem, CategorySlice, LatestEntry, SummaryCard } from '../models/dashboard.model';

const SHIELD_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#3b82f6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>';
const TREND_UP_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#22c55e"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>';
const TREND_DOWN_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#ef4444"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 17h8m0 0v-8m0 8l-8-8-4 4-6-6" /></svg>';

const INVEST_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#f59e0b"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>';

export const INITIAL_SUMMARY_CARDS: readonly SummaryCard[] = [
  { label: 'Saldo Total', value: 'R$ 0,00', bgClass: 'bg-secondary-light', icon: SHIELD_ICON },
  { label: 'Entradas (Mês)', value: 'R$ 0,00', bgClass: 'bg-success-light', icon: TREND_UP_ICON },
  { label: 'Saídas (Mês)', value: 'R$ 0,00', bgClass: 'bg-danger-light', icon: TREND_DOWN_ICON },
  { label: 'Carteira de Investimentos', value: 'R$ 0,00', bgClass: 'bg-gold-light', icon: INVEST_ICON },
];

export const INITIAL_CATEGORY_SLICES: readonly CategorySlice[] = [];

export const INITIAL_BUDGET_VS_ACTUAL: readonly BudgetVsActualItem[] = [];

export const INITIAL_LATEST_ENTRIES: readonly LatestEntry[] = [];
