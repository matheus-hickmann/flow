import type { BudgetVsActualItem, CategorySlice, LatestEntry, SummaryCard } from '../models/dashboard.model';

const SHIELD_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#3b82f6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>';
const TREND_UP_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#22c55e"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>';
const TREND_DOWN_ICON =
  '<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="#ef4444"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 17h8m0 0v-8m0 8l-8-8-4 4-6-6" /></svg>';

export const INITIAL_SUMMARY_CARDS: readonly SummaryCard[] = [
  { label: 'Saldo Total', value: 'R$ 12.450,00', bgClass: 'bg-secondary-light', icon: SHIELD_ICON },
  { label: 'Receitas (Mês)', value: 'R$ 5.200,00', bgClass: 'bg-success-light', icon: TREND_UP_ICON },
  { label: 'Despesas (Mês)', value: 'R$ 2.150,00', bgClass: 'bg-danger-light', icon: TREND_DOWN_ICON },
];

export const INITIAL_CATEGORY_SLICES: readonly CategorySlice[] = [
  { name: 'Alimentação', percent: 30, color: '#1e3a5f' },
  { name: 'Moradia', percent: 50, color: '#f97316' },
  { name: 'Transporte', percent: 10, color: '#22c55e' },
  { name: 'Outros', percent: 10, color: '#cbd5e1' },
];

export const INITIAL_BUDGET_VS_ACTUAL: readonly BudgetVsActualItem[] = [
  { category: 'Alimentação', budgeted: 1500, actual: 1650 },
  { category: 'Moradia', budgeted: 2500, actual: 3200 },
  { category: 'Transporte', budgeted: 800, actual: 950 },
  { category: 'Saúde', budgeted: 600, actual: 420 },
  { category: 'Educação', budgeted: 400, actual: 400 },
  { category: 'Lazer', budgeted: 500, actual: 680 },
  { category: 'Vestuário', budgeted: 300, actual: 250 },
  { category: 'Outros', budgeted: 400, actual: 520 },
];

export const INITIAL_LATEST_ENTRIES: readonly LatestEntry[] = [
  {
    id: '1',
    date: '10/02',
    description: 'Supermercado Extra',
    category: 'Alimentação',
    account: 'Itaú',
    value: '- R$ 350,00',
    isIncome: false,
  },
  {
    id: '2',
    date: '09/02',
    description: 'Salário Mensal',
    category: 'Renda',
    account: 'Santander',
    value: '+ R$ 5.000,00',
    isIncome: true,
  },
];
