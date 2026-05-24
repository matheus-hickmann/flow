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
import { ImportModalComponent } from '../../features/import';

// Paleta pastel para mapear contas que não possuam cor própria
const FLOW_PASTEL = ['#dcd2ec', '#f4d8c7', '#cce8d6', '#f3e9b9', '#cfdfeb', '#edd1d6', '#cdd9be'];
const WEEKDAYS = ['domingo', 'segunda', 'terça', 'quarta', 'quinta', 'sexta', 'sábado'];

interface AccountMeta {
  readonly name: string;
  readonly color: string;
  readonly balance: number;
  readonly movements: number;
}

interface DayGroup {
  readonly dayKey: string;
  readonly day: string;
  readonly weekday: string;
  readonly total: number;
  readonly rows: readonly Transaction[];
}

@Component({
  selector: 'app-transactions',
  standalone: true,
  imports: [CurrencyBrlPipe, DropdownComponent, ImportModalComponent],
  templateUrl: './transactions.component.html',
})
export class TransactionsComponent implements OnInit {
  protected readonly ariaLabelMonthPrevious = ARIA_LABELS.MONTH_PREVIOUS;
  protected readonly ariaLabelMonthNext = ARIA_LABELS.MONTH_NEXT;
  protected readonly categoriesList = [...TRANSACTION_CATEGORIES];

  private readonly accountService = inject(AccountService);
  private readonly transactionService = inject(TransactionService);
  private readonly env = inject(ENVIRONMENT);

  protected readonly accountsSignal = signal<Account[]>([]);
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
  private readonly filterAccountSignal = signal('');   // mantido p/ compatibilidade
  private readonly filterTypeSignal = signal<'all' | 'income' | 'expense'>('all');
  private readonly currentPageSignal = signal(0);

  // Conta selecionada no rail esquerdo
  private readonly selectedAccountSignal = signal<string>('');
  readonly selectedAccount = this.selectedAccountSignal.asReadonly();

  readonly showImportModal = signal(false);

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
    const acc = this.selectedAccountSignal() || this.filterAccountSignal().trim();
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

  // ─── Rail esquerdo: metadados por conta ────────────────────────────
  readonly accountsWithMeta = computed<AccountMeta[]>(() => {
    const accs = this.accountsSignal();
    const monthKey = this.monthKey();
    const txs = this.transactionsSignal().filter((t) => t.date.slice(0, 7) === monthKey);
    if (accs.length > 0) {
      return accs.map((a, i) => ({
        name: a.name,
        color: a.color || FLOW_PASTEL[i % FLOW_PASTEL.length],
        balance: Number(a.balance) || 0,
        movements: txs.filter((t) => t.account === a.name).length,
      }));
    }
    // fallback baseado em transações
    const names = [...new Set(txs.map((t) => t.account))];
    return names.map((name, i) => ({
      name,
      color: FLOW_PASTEL[i % FLOW_PASTEL.length],
      balance: 0,
      movements: txs.filter((t) => t.account === name).length,
    }));
  });

  readonly allAccountsBalance = computed(() => {
    const sum = this.accountsWithMeta().reduce((s, a) => s + a.balance, 0);
    return this.formatBRL(sum);
  });

  // ─── Agrupamento por dia ───────────────────────────────────────────
  readonly groupedByDay = computed<DayGroup[]>(() => {
    const rows = this.filteredByMonth();
    const buckets: Record<string, Transaction[]> = {};
    rows.forEach((r) => {
      const key = r.date.slice(0, 10);
      (buckets[key] = buckets[key] || []).push(r);
    });
    return Object.entries(buckets)
      .sort((a, b) => (a[0] < b[0] ? 1 : -1))
      .map(([dayKey, rs]) => {
        const [y, m, d] = dayKey.split('-').map(Number);
        const date = new Date(y, m - 1, d);
        const total = rs.reduce((s, t) => s + (t.isIncome ? t.value : -t.value), 0);
        return {
          dayKey,
          day: String(d).padStart(2, '0'),
          weekday: `${WEEKDAYS[date.getDay()]} · ${MONTH_LABELS[m - 1].toLowerCase()}`,
          total,
          rows: rs.sort((a, b) => (a.id < b.id ? 1 : -1)),
        };
      });
  });

  // ─── Cor da conta (pelo modelo Account ou fallback hash) ───────────
  accountColor(name: string): string {
    const meta = this.accountsWithMeta().find((a) => a.name === name);
    if (meta) return meta.color;
    let h = 0;
    for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) >>> 0;
    return FLOW_PASTEL[h % FLOW_PASTEL.length];
  }

  formatBRL(v: number): string {
    const abs = Math.abs(v);
    const s = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(abs);
    return v < 0 ? '−' + s : s;
  }

  // ─── Lifecycle ─────────────────────────────────────────────────────
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

  // ─── Mês / filtros ─────────────────────────────────────────────────
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

  setSelectedAccount(name: string): void {
    this.selectedAccountSignal.set(name);
    this.currentPageSignal.set(0);
  }

  setFilterDescription(event: Event): void {
    this.filterDescriptionSignal.set((event.target as HTMLInputElement).value);
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

  openImport(): void {
    this.showImportModal.set(true);
  }

  onImportDone(): void {
    this.showImportModal.set(false);
    this.refreshTransactions();
  }

  private getInitialDate(): { year: number; month: number } {
    const now = new Date();
    return { year: now.getFullYear(), month: now.getMonth() };
  }
}
