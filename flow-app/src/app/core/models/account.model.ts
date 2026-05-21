export interface Account {
  readonly id: string;
  readonly code: string;
  readonly name: string;
  readonly type: string;
  readonly balance: number;
  readonly color: string;
  readonly isSystem?: boolean;
  readonly investment?: boolean;
  readonly annualRate?: number;
  readonly brand?: string;
  readonly limit?: number;
  readonly closingDay?: number;
  readonly dueDay?: number;
}

export interface CreateAccountPayload {
  readonly name: string;
  readonly initialBalance: number;
  readonly color: string;
  readonly system?: boolean;
  readonly investment?: boolean;
  readonly annualRate?: number;
  readonly brand?: string;
  readonly limit?: number;
  readonly closingDay?: number;
  readonly dueDay?: number;
}

export interface AdjustBalancePayload {
  readonly newBalance: number;
}

export interface RenameAccountPayload {
  readonly name: string;
  readonly color?: string;
  readonly investment?: boolean;
  readonly annualRate?: number;
  readonly brand?: string;
  readonly limit?: number;
  readonly closingDay?: number;
  readonly dueDay?: number;
}
