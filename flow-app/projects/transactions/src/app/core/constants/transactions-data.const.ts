import type { Transaction } from '../models/transaction.model';

export const MOCK_TRANSACTIONS: readonly Transaction[] = [
  { id: '1', date: '2026-02-10', description: 'Supermercado Extra', category: 'Alimentação', account: 'Itaú', value: 350, isIncome: false },
  { id: '2', date: '2026-02-09', description: 'Salário Mensal', category: 'Salário', account: 'Santander', value: 5000, isIncome: true },
  { id: '3', date: '2026-02-08', description: 'Aluguel', category: 'Moradia', account: 'Itaú', value: 1800, isIncome: false },
  { id: '4', date: '2026-02-07', description: 'Uber', category: 'Transporte', account: 'Nubank', value: 45, isIncome: false },
  { id: '5', date: '2026-02-05', description: 'Freelance Projeto X', category: 'Freelance', account: 'Santander', value: 1200, isIncome: true },
  { id: '6', date: '2026-02-04', description: 'Restaurante', category: 'Alimentação', account: 'Nubank', value: 89, isIncome: false },
  { id: '7', date: '2026-02-03', description: 'Conta de luz', category: 'Moradia', account: 'Itaú', value: 220, isIncome: false },
  { id: '8', date: '2026-02-02', description: 'Dividendos', category: 'Investimentos', account: 'Santander', value: 150, isIncome: true },
  { id: '9', date: '2026-02-01', description: 'Academia', category: 'Saúde', account: 'Nubank', value: 99, isIncome: false },
  { id: '10', date: '2026-01-28', description: 'Supermercado', category: 'Alimentação', account: 'Itaú', value: 420, isIncome: false },
  { id: '11', date: '2026-01-25', description: 'Salário Mensal', category: 'Salário', account: 'Santander', value: 5000, isIncome: true },
  { id: '12', date: '2026-01-20', description: 'Gasolina', category: 'Transporte', account: 'Itaú', value: 280, isIncome: false },
  { id: '13', date: '2026-01-15', description: 'Aluguel', category: 'Moradia', account: 'Itaú', value: 1800, isIncome: false },
  { id: '14', date: '2026-01-10', description: 'Netflix', category: 'Lazer', account: 'Nubank', value: 55, isIncome: false },
  { id: '15', date: '2026-01-05', description: 'Consulta médica', category: 'Saúde', account: 'Dinheiro', value: 200, isIncome: false },
];

export const TRANSACTION_CATEGORIES = [
  'Alimentação', 'Moradia', 'Transporte', 'Saúde', 'Educação', 'Lazer', 'Vestuário',
  'Salário', 'Freelance', 'Investimentos', 'Aluguel recebido', 'Outros',
] as const;

export const TRANSACTION_ACCOUNTS = ['Itaú', 'Santander', 'Nubank', 'Dinheiro'] as const;

export const TRANSACTIONS_PAGE_SIZE = 10;
