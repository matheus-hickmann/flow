export interface Debt {
  readonly id: string;
  readonly name: string;
  readonly amount: number;
  readonly remaining: number;
  readonly type: 'TO_PAY' | 'TO_RECEIVE';
  readonly counterparty: string;
  readonly dueDate?: string;
  readonly notes?: string;
  readonly status: 'ACTIVE' | 'SETTLED';
  readonly createdAt: string;
}

export interface CreateDebtPayload {
  readonly name: string;
  readonly amount: number;
  readonly type: 'TO_PAY' | 'TO_RECEIVE';
  readonly counterparty: string;
  readonly dueDate?: string;
  readonly notes?: string;
}

export interface DebtPaymentPayload {
  readonly amount: number;
}
