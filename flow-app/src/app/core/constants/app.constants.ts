export const ROUTES = {
  DASHBOARD: '',
  TRANSACTIONS: 'transacoes',
  ACCOUNTS: 'contas',
  REPORTS: 'relatorios',
  CATEGORIES: 'categorias',
} as const;

export const ARIA_LABELS = {
  ADD_ENTRY: 'Adicionar lançamento',
  FAB_EXPENSE: 'Despesa',
  FAB_INCOME: 'Receita',
  FAB_PLANNING: 'Planejamento',
  FAB_TOGGLE: 'Abrir ou fechar menu de adicionar',
  MONTH_PREVIOUS: 'Mês anterior',
  MONTH_NEXT: 'Próximo mês',
} as const;

export const MONTH_LABELS = [
  'JAN', 'FEV', 'MAR', 'ABR', 'MAI', 'JUN',
  'JUL', 'AGO', 'SET', 'OUT', 'NOV', 'DEZ',
] as const;

export const FAB_OPTIONS = [
  { id: 'expense', label: 'Despesa' },
  { id: 'income', label: 'Receita' },
  { id: 'transfer', label: 'Transferência' },
  { id: 'planning', label: 'Planejamento' },
] as const;
