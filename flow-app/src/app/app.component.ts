import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { Meta } from '@angular/platform-browser';

import { ARIA_LABELS, FAB_OPTIONS } from './core/constants/app.constants';
import type { Account } from './core';
import { AccountService, AuthService, DashboardService, ENVIRONMENT, PlanningService, TransactionService } from './core';
import type { BudgetResponse } from './core/services/planning.service';
import { HeaderComponent, FabMenuComponent, type FabOptionId } from './layout';
import {
  ExpenseEntryModalComponent,
  IncomeEntryModalComponent,
  PlanningEntryModalComponent,
  TransferEntryModalComponent,
  type PlanningSubmitPayload,
  type TransferSubmitPayload,
} from './features/entries';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterOutlet,
    HeaderComponent,
    FabMenuComponent,
    ExpenseEntryModalComponent,
    IncomeEntryModalComponent,
    PlanningEntryModalComponent,
    TransferEntryModalComponent,
  ],
  templateUrl: './app.component.html',
})
export class AppComponent implements OnInit {
  protected readonly ariaLabelFabToggle = ARIA_LABELS.FAB_TOGGLE;
  protected readonly fabOptions = FAB_OPTIONS;

  private readonly accountService = inject(AccountService);
  private readonly transactionService = inject(TransactionService);
  private readonly planningService = inject(PlanningService);
  private readonly dashboardService = inject(DashboardService);
  private readonly env = inject(ENVIRONMENT);
  private readonly meta = inject(Meta);
  protected readonly authService = inject(AuthService);

  private readonly isFabOpenSignal = signal(false);
  readonly isFabOpen = this.isFabOpenSignal.asReadonly();

  private readonly accountsSignal = signal<Account[]>([]);
  private readonly budgetsSignal = signal<BudgetResponse[]>([]);

  /** Account names for income/expense modals. */
  readonly accountNamesForModals = computed(() => this.accountsSignal().map((a) => a.name));
  /** Full accounts for transfer modal. */
  readonly accountsForTransfer = computed(() => this.accountsSignal());
  /** Budget limits for income/expense modals. */
  readonly budgetsForModals = this.budgetsSignal.asReadonly();

  ngOnInit(): void {
    this.authService.refreshUserFromBackend();
    if (this.env.appUrl) {
      const ogImage = `${this.env.appUrl}/assets/og.svg`;
      this.meta.updateTag({ property: 'og:image', content: ogImage });
      this.meta.updateTag({ name: 'twitter:image', content: ogImage });
    }
  }

  toggleFab(): void {
    this.isFabOpenSignal.update((open) => !open);
  }

  private readonly openModalSignal = signal<'expense' | 'income' | 'planning' | 'transfer' | null>(null);
  readonly openModal = this.openModalSignal.asReadonly();

  onFabOptionSelect(optionId: FabOptionId): void {
    this.openModalSignal.set(optionId);
    this.isFabOpenSignal.set(false);
    if ((optionId === 'income' || optionId === 'expense' || optionId === 'transfer') && this.env.apiUrl) {
      this.accountService.list().subscribe({
        next: (accounts) => {
          this.accountsSignal.set(accounts);
          if (optionId !== 'transfer') {
            this.planningService.listBudgets().subscribe({
              next: (budgets) => this.budgetsSignal.set(budgets),
            });
          }
        },
      });
    }
  }

  closeModal(): void {
    this.openModalSignal.set(null);
  }

  onExpenseSubmitted(payload: {
    description: string;
    value: number;
    category: string;
    account: string;
    date: string | null;
    budgetLimitId?: string;
  }): void {
    if (!this.env.apiUrl) {
      console.log('Despesa cadastrada (mock):', payload);
      this.closeModal();
      return;
    }
    const accounts = this.accountsSignal();
    const accountId = accounts.find((a) => a.name === payload.account)?.id;
    if (!accountId) {
      console.error('Conta não encontrada.');
      return;
    }
    this.transactionService
      .postTransaction({
        description: payload.description,
        category: payload.category,
        budgetLimitId: payload.budgetLimitId,
        entries: [
          { accountId, amount: payload.value, type: 'CREDIT' },
        ],
      })
      .subscribe({
        next: () => {
          this.transactionService.refresh();
          this.closeModal();
        },
        error: (err) => console.error('Erro ao cadastrar despesa:', err),
      });
  }

  onIncomeSubmitted(payload: {
    description: string;
    value: number;
    category: string;
    account: string;
    date: string | null;
    budgetLimitId?: string;
  }): void {
    if (!this.env.apiUrl) {
      console.log('Receita cadastrada (mock):', payload);
      this.closeModal();
      return;
    }
    const accounts = this.accountsSignal();
    const accountId = accounts.find((a) => a.name === payload.account)?.id;
    if (!accountId) {
      console.error('Conta não encontrada.');
      return;
    }
    this.transactionService
      .postTransaction({
        description: payload.description,
        category: payload.category,
        entries: [
          { accountId, amount: payload.value, type: 'DEBIT' },
        ],
      })
      .subscribe({
        next: () => {
          this.transactionService.refresh();
          this.closeModal();
        },
        error: (err) => console.error('Erro ao cadastrar receita:', err),
      });
  }

  onTransferSubmitted(payload: TransferSubmitPayload): void {
    if (!this.env.apiUrl) {
      console.log('Transferência (mock):', payload);
      this.closeModal();
      return;
    }
    this.transactionService
      .postTransaction({
        description: payload.description,
        category: 'Transferência',
        entries: [
          { accountId: payload.fromAccountId, amount: payload.amount, type: 'CREDIT' },
          { accountId: payload.toAccountId, amount: payload.amount, type: 'DEBIT' },
        ],
      })
      .subscribe({
        next: () => {
          this.transactionService.refresh();
          this.closeModal();
        },
        error: (err) => console.error('Erro ao registrar transferência:', err),
      });
  }

  onPlanningSubmitted(payload: PlanningSubmitPayload): void {
    this.planningService.submitPlanning(payload).subscribe({
      next: () => {
        if (payload.type === 'limit') {
          this.dashboardService.refresh();
        }
        this.closeModal();
      },
      error: () => this.closeModal(),
    });
  }
}
