import { Component, computed, inject, OnInit, signal } from '@angular/core';

import { ARIA_LABELS, MONTH_LABELS } from '../../core/constants/app.constants';
import {
  MOCK_TRANSACTIONS,
  TRANSACTION_ACCOUNTS,
  TRANSACTION_CATEGORIES,
  TRANSACTIONS_PAGE_SIZE,
} from '../../core/constants/transactions-data.const';
import type { Account, Transaction } from '../../core';
import { AccountService, ENVIRONMENT, TransactionService } from '../../core';
import { mapTransactionListItemsToTransactions } from '../../core/utils/transaction-mapper';
import { CurrencyBrlPipe, DropdownComponent } from '../../shared';

@Component({
  selector: 'app-transactions',
  standalone: true,
  imports: [CurrencyBrlPipe, DropdownComponent],
  templateUrl: './transactions.component.html',
})
export class TransactionsComponent implements OnInit {
  protected readonly ariaLabelMonthPrevious = ARIA_LABELS.MONTH_PREVIOUS;
  protected readonly ariaLabelMonthNext = ARIA_LABELS.MONTH_NEXT;
  protected readonly categoriesList = [...TRANSACTION_CATEGORIES];

  private readonly accountService = inject(AccountService);
  private readonly transactionService = inject(TransactionService);
  private readonly env = inject(ENVIRONMENT);

  private readonly accountsSignal = signal<Account[]>([]);
  private readonly transactionsSignal = signal<Transaction[]>([]);
  private readonly loadingSignal = signal(false);

  protected readonly accountsList = computed(() => {
    const accounts = this.accountsSignal();
    if (accounts.length > 0) return accounts.map((a) => a.name);
    return [...TRANSACTION_ACCOUNTS];
  });

  private readonly currentDateSignal = signal<{ year: number; month: number }>(this.getInitialDate());
  private readonly filterDescriptionSignal = signal('');
  private readonly filterCategorySignal = signal('');
  private readonly filterAccountSignal = signal('');
  private readonly filterTypeSignal = signal<'all' | 'income' | 'expense'>('all');
  private readonly currentPageSignal = signal(0);

  readonly monthLabel = computed(() => {
    const { year, month } = this.currentDateSignal();
    return `${MONTH_LABELS[month]} ${year}`;
  });
  readonly monthKey = computed(() => {
    const { year, month } = this.currentDateSignal();
    return `${year}-${String(month + 1).padStart(2, '0')}`;
  });
  readonly filterDescription = this.filterDescriptionSignal.asReadonly();
  readonly filterCategory = this.filterCategorySignal.asReadonly();
  readonly filterAccount = this.filterAccountSignal.asReadonly();
  readonly filterType = this.filterTypeSignal.asReadonly();
  readonly currentPage = this.currentPageSignal.asReadonly();
  readonly loading = this.loadingSignal.asReadonly();

  readonly filteredByMonth = computed(() => {
    const transactions = this.transactionsSignal();
    const monthKey = this.monthKey();
    const desc = this.filterDescriptionSignal().toLowerCase().trim();
    const cat = this.filterCategorySignal().trim();
    const acc = this.filterAccountSignal().trim();
    const type = this.filterTypeSignal();
    return transactions.filter((tx) => {
      if (tx.date.slice(0, 7) !== monthKey) return false;
      if (desc && !tx.description.toLowerCase().includes(desc)) return false;
      if (cat && tx.category !== cat) return false;
      if (acc && tx.account !== acc) return false;
      if (type === 'income' && !tx.isIncome) return false;
      if (type === 'expense' && tx.isIncome) return false;
      return true;
    });
  });

  readonly totalIncome = computed(() =>
    this.filteredByMonth().filter((t) => t.isIncome).reduce((s, t) => s + t.value, 0),
  );
  readonly totalExpense = computed(() =>
    this.filteredByMonth().filter((t) => !t.isIncome).reduce((s, t) => s + t.value, 0),
  );
  readonly balance = computed(() =>
    this.filteredByMonth().reduce((s, t) => s + (t.isIncome ? t.value : -t.value), 0),
  );

  readonly totalPages = computed(() =>
    Math.max(1, Math.ceil(this.filteredByMonth().length / TRANSACTIONS_PAGE_SIZE)),
  );
  readonly fromIndex = computed(() => this.currentPageSignal() * TRANSACTIONS_PAGE_SIZE);
  readonly toIndex = computed(() =>
    Math.min(this.fromIndex() + TRANSACTIONS_PAGE_SIZE, this.filteredByMonth().length),
  );
  readonly paginatedTransactions = computed(() => {
    const list = this.filteredByMonth();
    const start = this.currentPageSignal() * TRANSACTIONS_PAGE_SIZE;
    return list.slice(start, start + TRANSACTIONS_PAGE_SIZE);
  });

  ngOnInit(): void {
    if (!this.env.apiUrl) {
      this.transactionsSignal.set([...MOCK_TRANSACTIONS]);
      return;
    }
    this.loadingSignal.set(true);
    this.accountService.list().subscribe({
      next: (accounts) => {
        this.accountsSignal.set(accounts);
        this.loadTransactionList(accounts);
      },
      error: () => {
        this.transactionsSignal.set([...MOCK_TRANSACTIONS]);
        this.loadingSignal.set(false);
      },
    });
    this.transactionService.refresh$.subscribe(() => this.refreshTransactions());
  }

  private loadTransactionList(accounts: Account[]): void {
    this.transactionService.list(200).subscribe({
      next: (items) => {
        this.transactionsSignal.set(mapTransactionListItemsToTransactions(items, accounts));
        this.loadingSignal.set(false);
      },
      error: () => {
        this.transactionsSignal.set([...MOCK_TRANSACTIONS]);
        this.loadingSignal.set(false);
      },
    });
  }

  refreshTransactions(): void {
    const accounts = this.accountsSignal();
    if (!this.env.apiUrl || accounts.length === 0) return;
    this.loadTransactionList(accounts);
  }

  previousMonth(): void {
    this.currentDateSignal.update(({ year, month }) => {
      if (month === 0) return { year: year - 1, month: 11 };
      return { year, month: month - 1 };
    });
    this.currentPageSignal.set(0);
  }

  nextMonth(): void {
    this.currentDateSignal.update(({ year, month }) => {
      if (month === 11) return { year: year + 1, month: 0 };
      return { year, month: month + 1 };
    });
    this.currentPageSignal.set(0);
  }

  setFilterDescription(event: Event): void {
    const value = (event.target as HTMLInputElement).value;
    this.filterDescriptionSignal.set(value);
    this.currentPageSignal.set(0);
  }

  setFilterCategoryValue(value: string): void {
    this.filterCategorySignal.set(value);
    this.currentPageSignal.set(0);
  }

  setFilterAccountValue(value: string): void {
    this.filterAccountSignal.set(value);
    this.currentPageSignal.set(0);
  }

  setFilterTypeValue(value: string): void {
    this.filterTypeSignal.set(value as 'all' | 'income' | 'expense');
    this.currentPageSignal.set(0);
  }

  prevPage(): void {
    this.currentPageSignal.update((p) => Math.max(0, p - 1));
  }

  nextPage(): void {
    this.currentPageSignal.update((p) => Math.min(this.totalPages() - 1, p + 1));
  }

  formatDate(iso: string): string {
    const [y, m, d] = iso.split('-');
    return `${d}/${m}`;
  }

  private getInitialDate(): { year: number; month: number } {
    const now = new Date();
    return { year: now.getFullYear(), month: now.getMonth() };
  }
}
