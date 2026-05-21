import { Component, input, inject, output } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ModalComponent } from '../../shared';
import type { Account } from '../../core';

export interface TransferSubmitPayload {
  readonly description: string;
  readonly fromAccountId: string;
  readonly toAccountId: string;
  readonly amount: number;
  readonly date: string | null;
}

@Component({
  selector: 'app-transfer-entry-modal',
  standalone: true,
  imports: [ReactiveFormsModule, ModalComponent],
  templateUrl: './transfer-entry-modal.component.html',
})
export class TransferEntryModalComponent {
  readonly submitted = output<TransferSubmitPayload>();
  readonly closed = output<void>();

  readonly accountsInput = input<Account[]>([]);

  private readonly fb = inject(FormBuilder);
  form: FormGroup = this.fb.group({
    description: ['Transferência', [Validators.required, Validators.maxLength(500)]],
    fromAccountId: ['', Validators.required],
    toAccountId: ['', Validators.required],
    amount: [null as number | null, [Validators.required, Validators.min(0.01)]],
    date: [this.todayString()],
  });

  protected get accounts(): Account[] {
    return this.accountsInput();
  }

  protected get isDifferentAccounts(): boolean {
    const v = this.form.value;
    return !v.fromAccountId || !v.toAccountId || v.fromAccountId !== v.toAccountId;
  }

  onClose(): void {
    this.closed.emit();
  }

  onSubmit(): void {
    if (this.form.invalid || !this.isDifferentAccounts) return;
    const v = this.form.value;
    this.submitted.emit({
      description: v.description,
      fromAccountId: v.fromAccountId,
      toAccountId: v.toAccountId,
      amount: Number(v.amount),
      date: v.date ?? null,
    });
    this.closed.emit();
  }

  private todayString(): string {
    return new Date().toISOString().slice(0, 10);
  }
}
