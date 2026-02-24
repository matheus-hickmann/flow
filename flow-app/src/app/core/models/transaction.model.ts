export interface Transaction {
  readonly id: string;
  readonly date: string;
  readonly description: string;
  readonly category: string;
  readonly account: string;
  readonly value: number;
  readonly isIncome: boolean;
}

export interface TransactionFilters {
  readonly description: string;
  readonly category: string;
  readonly account: string;
  readonly type: 'all' | 'income' | 'expense';
}

export const DEFAULT_TRANSACTION_FILTERS: TransactionFilters = {
  description: '',
  category: '',
  account: '',
  type: 'all',
};
