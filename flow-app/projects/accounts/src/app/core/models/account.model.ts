export interface Account {
  readonly id: number;
  readonly code: string;
  readonly name: string;
  readonly type: string;
  readonly balance: number;
  readonly color: string;
}

export interface CreateAccountPayload {
  readonly name: string;
  readonly initialBalance: number;
  readonly color: string;
}

export interface AdjustBalancePayload {
  readonly newBalance: number;
}
