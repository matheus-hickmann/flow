import { Component, inject, output } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ModalComponent } from '../../shared';

@Component({
  selector: 'app-expense-entry-modal',
  standalone: true,
  imports: [ReactiveFormsModule, ModalComponent],
  templateUrl: './expense-entry-modal.component.html',
})
export class ExpenseEntryModalComponent {
  readonly submitted = output<{ description: string; value: number; category: string; account: string; date: string | null }>();
  readonly closed = output<void>();

  private readonly fb = inject(FormBuilder);
  form: FormGroup = this.fb.group({
    description: ['', [Validators.required, Validators.maxLength(500)]],
    value: [null as number | null, [Validators.required, Validators.min(0.01)]],
    category: ['', Validators.required],
    account: ['', Validators.required],
    date: [this.todayString()],
  });
  protected readonly categories = ['Alimentação', 'Moradia', 'Transporte', 'Saúde', 'Educação', 'Lazer', 'Vestuário', 'Outros'];
  protected readonly accounts = ['Itaú', 'Santander', 'Nubank', 'Dinheiro'];

  onClose(): void {
    this.closed.emit();
  }

  onSubmit(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.submitted.emit({
      description: v.description,
      value: Number(v.value),
      category: v.category,
      account: v.account,
      date: v.date ?? null,
    });
    this.closed.emit();
  }

  private todayString(): string {
    const t = new Date();
    return t.toISOString().slice(0, 10);
  }
}
