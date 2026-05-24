import { Component, inject, OnInit, signal } from '@angular/core';

import type { Account, CreateAccountPayload, RenameAccountPayload } from '../../core';
import { AccountService } from '../../core';
import { ModalComponent, CurrencyBrlPipe } from '../../shared';

const DEFAULT_COLOR = '#3b82f6';

export const CARD_BRANDS = ['Visa', 'Mastercard', 'Elo', 'American Express', 'Hipercard', 'Outro'] as const;
export type AccountKind = 'regular' | 'investment' | 'credit_card';

@Component({
  selector: 'app-accounts',
  standalone: true,
  imports: [ModalComponent, CurrencyBrlPipe],
  templateUrl: './accounts.component.html',
})
export class AccountsComponent implements OnInit {
  private readonly accountService = inject(AccountService);

  readonly accounts = signal<Account[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly cardBrands = CARD_BRANDS;

  readonly showCreateModal = signal(false);
  readonly createName = signal('');
  readonly createInitialBalance = signal<number>(0);
  readonly createColor = signal(DEFAULT_COLOR);
  readonly createKind = signal<AccountKind>('regular');
  readonly createIsInvestment = signal(false);
  readonly createAnnualRate = signal<number>(0);
  readonly createBrand = signal<string>(CARD_BRANDS[0]);
  readonly createLimit = signal<number>(0);
  readonly createClosingDay = signal<number>(1);
  readonly createDueDay = signal<number>(10);
  readonly createNameError = signal<string | null>(null);

  readonly showAdjustModal = signal<Account | null>(null);
  readonly adjustNewBalance = signal<number | null>(null);
  readonly adjustError = signal<string | null>(null);

  readonly showEditModal = signal<Account | null>(null);
  readonly editName = signal('');
  readonly editColor = signal(DEFAULT_COLOR);
  readonly editKind = signal<AccountKind>('regular');
  readonly editAnnualRate = signal<number>(0);
  readonly editBrand = signal<string>(CARD_BRANDS[0]);
  readonly editLimit = signal<number>(0);
  readonly editClosingDay = signal<number>(1);
  readonly editDueDay = signal<number>(10);
  readonly editError = signal<string | null>(null);

  readonly showDeleteConfirm = signal<Account | null>(null);

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
    this.createKind.set('regular');
    this.createIsInvestment.set(false);
    this.createAnnualRate.set(0);
    this.createBrand.set(CARD_BRANDS[0]);
    this.createLimit.set(0);
    this.createClosingDay.set(1);
    this.createDueDay.set(10);
    this.createNameError.set(null);
    this.showCreateModal.set(true);
  }

  setCreateKind(kind: AccountKind): void {
    this.createKind.set(kind);
    this.createIsInvestment.set(kind === 'investment');
    if (kind !== 'investment') this.createAnnualRate.set(0);
  }

  setCreateIsInvestment(value: boolean): void {
    this.createIsInvestment.set(value);
    if (!value) this.createAnnualRate.set(0);
  }

  setCreateAnnualRate(event: Event): void {
    const raw = (event.target as HTMLInputElement).value;
    const n = parseFloat(raw);
    this.createAnnualRate.set(Number.isNaN(n) ? 0 : n);
  }

  setCreateBrand(event: Event): void {
    this.createBrand.set((event.target as HTMLSelectElement).value);
  }

  setCreateLimit(event: Event): void {
    const n = parseFloat((event.target as HTMLInputElement).value);
    this.createLimit.set(Number.isNaN(n) ? 0 : n);
  }

  setCreateClosingDay(event: Event): void {
    const n = parseInt((event.target as HTMLInputElement).value, 10);
    this.createClosingDay.set(Number.isNaN(n) ? 1 : Math.min(31, Math.max(1, n)));
  }

  setCreateDueDay(event: Event): void {
    const n = parseInt((event.target as HTMLInputElement).value, 10);
    this.createDueDay.set(Number.isNaN(n) ? 1 : Math.min(31, Math.max(1, n)));
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
    const kind = this.createKind();
    const isCreditCard = kind === 'credit_card';
    const isInvestment = kind === 'investment';
    const payload: CreateAccountPayload = {
      name,
      initialBalance: isCreditCard ? 0 : this.createInitialBalance(),
      color: this.createColor(),
      investment: isInvestment,
      annualRate: isInvestment ? this.createAnnualRate() : undefined,
      brand: isCreditCard ? this.createBrand() : undefined,
      limit: isCreditCard ? this.createLimit() : undefined,
      closingDay: isCreditCard ? this.createClosingDay() : undefined,
      dueDay: isCreditCard ? this.createDueDay() : undefined,
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

  openEditModal(account: Account): void {
    const kind: AccountKind = account.type === 'CREDIT_CARD'
      ? 'credit_card'
      : account.investment ? 'investment' : 'regular';
    this.showEditModal.set(account);
    this.editName.set(account.name);
    this.editColor.set(account.color ?? DEFAULT_COLOR);
    this.editKind.set(kind);
    this.editAnnualRate.set(account.annualRate ?? 0);
    this.editBrand.set(account.brand ?? CARD_BRANDS[0]);
    this.editLimit.set(account.limit ?? 0);
    this.editClosingDay.set(account.closingDay ?? 1);
    this.editDueDay.set(account.dueDay ?? 10);
    this.editError.set(null);
  }

  closeEditModal(): void {
    this.showEditModal.set(null);
  }

  setEditName(event: Event): void {
    this.editName.set((event.target as HTMLInputElement).value);
    this.editError.set(null);
  }

  setEditColor(event: Event): void {
    this.editColor.set((event.target as HTMLInputElement).value);
  }

  setEditColorHex(event: Event): void {
    const v = (event.target as HTMLInputElement).value.trim();
    if (/^#[0-9A-Fa-f]{6}$/.test(v)) this.editColor.set(v);
  }

  setEditKind(kind: AccountKind): void {
    this.editKind.set(kind);
    if (kind !== 'investment') this.editAnnualRate.set(0);
  }

  setEditAnnualRate(event: Event): void {
    const n = parseFloat((event.target as HTMLInputElement).value);
    this.editAnnualRate.set(Number.isNaN(n) ? 0 : n);
  }

  setEditBrand(event: Event): void {
    this.editBrand.set((event.target as HTMLSelectElement).value);
  }

  setEditLimit(event: Event): void {
    const n = parseFloat((event.target as HTMLInputElement).value);
    this.editLimit.set(Number.isNaN(n) ? 0 : n);
  }

  setEditClosingDay(event: Event): void {
    const n = parseInt((event.target as HTMLInputElement).value, 10);
    this.editClosingDay.set(Number.isNaN(n) ? 1 : Math.min(31, Math.max(1, n)));
  }

  setEditDueDay(event: Event): void {
    const n = parseInt((event.target as HTMLInputElement).value, 10);
    this.editDueDay.set(Number.isNaN(n) ? 1 : Math.min(31, Math.max(1, n)));
  }

  onSubmitEdit(): void {
    const account = this.showEditModal();
    const name = this.editName().trim();
    if (!account || !name) {
      this.editError.set('Informe o nome da conta.');
      return;
    }
    const kind = this.editKind();
    const isCreditCard = kind === 'credit_card';
    const isInvestment = kind === 'investment';
    const payload: RenameAccountPayload = {
      name,
      color: this.editColor(),
      investment: isInvestment,
      annualRate: isInvestment ? this.editAnnualRate() : undefined,
      brand: isCreditCard ? this.editBrand() : undefined,
      limit: isCreditCard ? this.editLimit() : undefined,
      closingDay: isCreditCard ? this.editClosingDay() : undefined,
      dueDay: isCreditCard ? this.editDueDay() : undefined,
    };
    this.accountService.rename(account.id, payload).subscribe({
      next: (updated) => {
        this.accounts.update((list) => list.map((a) => (a.id === updated.id ? updated : a)));
        this.closeEditModal();
      },
      error: () => {
        this.editError.set('Não foi possível salvar as alterações. Tente novamente.');
      },
    });
  }

  openDeleteConfirm(account: Account): void {
    this.showDeleteConfirm.set(account);
  }

  closeDeleteConfirm(): void {
    this.showDeleteConfirm.set(null);
  }

  onConfirmDelete(): void {
    const account = this.showDeleteConfirm();
    if (!account) return;
    this.accountService.delete(account.id).subscribe({
      next: () => {
        this.accounts.update((list) => list.filter((a) => a.id !== account.id));
        this.closeDeleteConfirm();
      },
      error: () => {
        this.closeDeleteConfirm();
      },
    });
  }

  toggleShared(account: Account): void {
    const newShared = !account.shared;
    this.accountService.update(account.id, { shared: newShared }).subscribe({
      next: (updated) => {
        this.accounts.update((list) => list.map((a) => (a.id === updated.id ? updated : a)));
      },
    });
  }

}
