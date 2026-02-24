import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { ENVIRONMENT } from './core/config';
import type { Account, CreateAccountPayload } from './core/models/account.model';
import { AccountService } from './core/services/account.service';
import { ModalComponent, CurrencyBrlPipe } from './shared';

const DEFAULT_COLOR = '#3b82f6';

const DEFAULT_ENV = {
  production: false,
  appName: 'Flow Accounts',
  defaultUserName: '',
  apiUrl: '',
};

@Component({
  selector: 'app-accounts',
  standalone: true,
  imports: [FormsModule, ModalComponent, CurrencyBrlPipe],
  templateUrl: './app.html',
  providers: [
    { provide: ENVIRONMENT, useValue: DEFAULT_ENV },
    AccountService,
  ],
})
export class App implements OnInit {
  private readonly accountService = inject(AccountService);

  readonly accounts = signal<Account[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly showCreateModal = signal(false);
  readonly createName = signal('');
  readonly createInitialBalance = signal<number>(0);
  readonly createColor = signal(DEFAULT_COLOR);
  readonly createNameError = signal<string | null>(null);

  readonly showAdjustModal = signal<Account | null>(null);
  readonly adjustNewBalance = signal<number | null>(null);
  readonly adjustError = signal<string | null>(null);

  ngOnInit(): void {
    this.loadAccounts();
  }

  loadAccounts(): void {
    this.loading.set(true);
    this.error.set(null);
    this.accountService.list().subscribe({
      next: (list) => {
        this.accounts.set(list);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err?.message ?? 'Erro ao carregar contas.');
        this.loading.set(false);
      },
    });
  }

  openCreateModal(): void {
    this.createName.set('');
    this.createInitialBalance.set(0);
    this.createColor.set(DEFAULT_COLOR);
    this.createNameError.set(null);
    this.showCreateModal.set(true);
  }

  closeCreateModal(): void {
    this.showCreateModal.set(false);
  }

  setCreateName(event: Event): void {
    this.createName.set((event.target as HTMLInputElement).value.trim());
    this.createNameError.set(null);
  }

  setCreateInitialBalance(event: Event): void {
    const raw = (event.target as HTMLInputElement).value;
    const n = parseFloat(raw);
    this.createInitialBalance.set(Number.isNaN(n) ? 0 : n);
  }

  setCreateColor(event: Event): void {
    this.createColor.set((event.target as HTMLInputElement).value);
  }

  setCreateColorHex(event: Event): void {
    const v = (event.target as HTMLInputElement).value.trim();
    if (/^#[0-9A-Fa-f]{6}$/.test(v)) this.createColor.set(v);
  }

  canSubmitCreate(): boolean {
    return this.createName().length > 0;
  }

  onSubmitCreate(): void {
    const name = this.createName().trim();
    if (!name) {
      this.createNameError.set('Informe o nome da conta.');
      return;
    }
    const payload: CreateAccountPayload = {
      name,
      initialBalance: this.createInitialBalance(),
      color: this.createColor(),
    };
    this.accountService.create(payload).subscribe({
      next: (created) => {
        this.accounts.update((list) => [...list, created]);
        this.closeCreateModal();
      },
      error: () => {
        this.createNameError.set('Não foi possível criar a conta. Tente novamente.');
      },
    });
  }

  openAdjustModal(account: Account): void {
    this.showAdjustModal.set(account);
    this.adjustNewBalance.set(account.balance);
    this.adjustError.set(null);
  }

  closeAdjustModal(): void {
    this.showAdjustModal.set(null);
  }

  setAdjustNewBalance(event: Event): void {
    const raw = (event.target as HTMLInputElement).value;
    const n = parseFloat(raw);
    this.adjustNewBalance.set(Number.isNaN(n) ? null : n);
  }

  onSubmitAdjust(): void {
    const account = this.showAdjustModal();
    const newBalance = this.adjustNewBalance();
    if (!account || newBalance === null) return;
    this.adjustError.set(null);
    this.accountService.adjustBalance(account.id, { newBalance }).subscribe({
      next: (updated) => {
        this.accounts.update((list) =>
          list.map((a) => (a.id === updated.id ? { ...a, balance: updated.balance } : a)),
        );
        this.closeAdjustModal();
      },
      error: () => {
        this.adjustError.set('Não foi possível ajustar o saldo. Tente novamente.');
      },
    });
  }
}
