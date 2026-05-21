/** API DTOs for ledger-service transactions */

export interface EntryResponseDto {
  readonly id: string;
  readonly accountId: string;
  readonly amount: number;
  readonly type: string;
}

export interface TransactionListItemDto {
  readonly id: string;
  readonly description: string;
  readonly timestamp: string;
  readonly referenceId: string;
  readonly category: string;
  readonly entries: EntryResponseDto[];
}

export interface PostTransactionRequestDto {
  readonly description: string;
  readonly referenceId?: string;
  readonly category?: string;
  readonly budgetLimitId?: string;
  readonly entries: { accountId: string; amount: number; type: string }[];
}
