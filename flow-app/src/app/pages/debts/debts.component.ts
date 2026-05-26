import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { DebtService } from '../../core/services/debt.service';
import type { Debt, CreateDebtPayload } from '../../core/models/debt.model';
import { CurrencyBrlPipe } from '../../shared';

@Component({
  selector: 'app-debts',
  standalone: true,
  imports: [FormsModule, CurrencyBrlPipe],
  templateUrl: './debts.component.html',
})
export class DebtsComponent implements OnInit {
  private readonly debtService = inject(DebtService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly debts = signal<Debt[]>([]);

  readonly showForm = signal(false);
  readonly formError = signal<string | null>(null);
  readonly formLoading = signal(false);

  readonly paymentDebt = signal<Debt | null>(null);
  readonly paymentAmount = signal('');
  readonly paymentError = signal<string | null>(null);
  readonly paymentLoading = signal(false);

  // New debt form fields
  readonly formName = signal('');
  readonly formAmount = signal('');
  readonly formType = signal<'TO_PAY' | 'TO_RECEIVE'>('TO_PAY');
  readonly formCounterparty = signal('');
  readonly formDueDate = signal('');
  readonly formNotes = signal('');

  readonly activeDebts = computed(() => this.debts().filter((d) => d.status === 'ACTIVE'));
  readonly settledDebts = computed(() => this.debts().filter((d) => d.status === 'SETTLED'));

  readonly totalToPayActive = computed(() =>
    this.activeDebts()
      .filter((d) => d.type === 'TO_PAY')
      .reduce((s, d) => s + d.remaining, 0),
  );
  readonly totalToReceiveActive = computed(() =>
    this.activeDebts()
      .filter((d) => d.type === 'TO_RECEIVE')
      .reduce((s, d) => s + d.remaining, 0),
  );
  readonly netBalance = computed(() => this.totalToReceiveActive() - this.totalToPayActive());

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.debtService.list().subscribe({
      next: (debts) => {
        this.debts.set(debts);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Erro ao carregar dívidas.');
        this.loading.set(false);
      },
    });
  }

  openForm(): void {
    this.formName.set('');
    this.formAmount.set('');
    this.formType.set('TO_PAY');
    this.formCounterparty.set('');
    this.formDueDate.set('');
    this.formNotes.set('');
    this.formError.set(null);
    this.showForm.set(true);
  }

  closeForm(): void {
    this.showForm.set(false);
  }

  submitForm(): void {
    const amount = parseFloat(String(this.formAmount()).replace(',', '.'));
    if (!this.formName().trim() || isNaN(amount) || amount <= 0) {
      this.formError.set('Nome e valor são obrigatórios.');
      return;
    }

    const payload: CreateDebtPayload = {
      name: this.formName().trim(),
      amount,
      type: this.formType(),
      counterparty: this.formCounterparty().trim(),
      dueDate: this.formDueDate() || undefined,
      notes: this.formNotes().trim() || undefined,
    };

    this.formLoading.set(true);
    this.formError.set(null);
    this.debtService.create(payload).subscribe({
      next: (res) => {
        this.formLoading.set(false);
        if (!res) {
          this.formError.set('Erro ao salvar dívida.');
          return;
        }
        this.showForm.set(false);
        this.load();
      },
      error: () => {
        this.formLoading.set(false);
        this.formError.set('Erro ao salvar dívida.');
      },
    });
  }

  openPayment(debt: Debt): void {
    this.paymentDebt.set(debt);
    this.paymentAmount.set('');
    this.paymentError.set(null);
  }

  closePayment(): void {
    this.paymentDebt.set(null);
  }

  submitPayment(): void {
    const debt = this.paymentDebt();
    if (!debt) return;
    const amount = parseFloat(String(this.paymentAmount()).replace(',', '.'));
    if (isNaN(amount) || amount <= 0) {
      this.paymentError.set('Informe um valor válido.');
      return;
    }
    if (amount > debt.remaining) {
      this.paymentError.set('Valor supera o saldo devedor.');
      return;
    }

    this.paymentLoading.set(true);
    this.paymentError.set(null);
    this.debtService.recordPayment(debt.id, { amount }).subscribe({
      next: (updated) => {
        this.paymentLoading.set(false);
        if (!updated) {
          this.paymentError.set('Erro ao registrar pagamento.');
          return;
        }
        this.paymentDebt.set(null);
        this.load();
      },
      error: () => {
        this.paymentLoading.set(false);
        this.paymentError.set('Erro ao registrar pagamento.');
      },
    });
  }

  deleteDebt(id: string): void {
    if (!confirm('Remover esta dívida?')) return;
    this.debtService.delete(id).subscribe({
      next: () => this.load(),
    });
  }

  progressPct(debt: Debt): number {
    if (debt.amount === 0) return 100;
    return Math.round(((debt.amount - debt.remaining) / debt.amount) * 100);
  }

  isOverdue(debt: Debt): boolean {
    if (!debt.dueDate || debt.status === 'SETTLED') return false;
    return new Date(debt.dueDate) < new Date();
  }

  formatDate(iso: string): string {
    const [year, month, day] = iso.split('-');
    return `${day}/${month}/${year}`;
  }
}
