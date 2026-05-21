import type { Account } from '../models/account.model';
import type { Transaction } from '../models/transaction.model';
import type { TransactionListItemDto } from '../models/transaction-api.model';

/** Parse backend timestamp (ISO or dd/MM/yyyy HH:mm:ss) to YYYY-MM-DD */
function toDateString(ts: string): string {
  if (ts.includes('T')) return ts.slice(0, 10);
  const [datePart] = ts.split(' ');
  if (datePart && datePart.includes('/')) {
    const [d, m, y] = datePart.split('/');
    return `${y}-${m!.padStart(2, '0')}-${d!.padStart(2, '0')}`;
  }
  return ts.slice(0, 10);
}

/** Parse amount (backend may send number or string "1.234,56") */
function toNumber(amount: number | string): number {
  if (typeof amount === 'number') return amount;
  const normalized = String(amount).trim().replace(/\./g, '').replace(',', '.');
  return Number(normalized) || 0;
}

export function mapTransactionListItemsToTransactions(
  items: TransactionListItemDto[],
  accounts: Account[],
): Transaction[] {
  const byId = new Map(accounts.map((a) => [a.id, a]));
  return items.map((tx) => {
    // Prefer the entry whose account belongs to the user; for single-entry income/expense
    // DEBIT = money in (income), CREDIT = money out (expense).
    // For transfers (two user accounts), the CREDIT entry is the origin account.
    const creditEntry = tx.entries.find((e) => e.type === 'CREDIT' && byId.has(e.accountId));
    const debitEntry = tx.entries.find((e) => e.type === 'DEBIT' && byId.has(e.accountId));
    const isTransfer = !!(creditEntry && debitEntry);
    const userEntry = isTransfer ? creditEntry : (debitEntry ?? creditEntry ?? tx.entries[0]);
    const amount = userEntry ? toNumber(userEntry.amount) : 0;
    const isIncome = !isTransfer && userEntry?.type === 'DEBIT';
    const accountName = userEntry ? (byId.get(userEntry.accountId)?.name ?? userEntry.accountId) : '';
    return {
      id: tx.id,
      date: toDateString(tx.timestamp),
      description: tx.description,
      category: tx.category || '',
      account: accountName,
      value: amount,
      isIncome,
    } satisfies Transaction;
  });
}
