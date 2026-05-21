import { Component, input, inject, output, OnInit, signal } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { CurrencyPipe } from '@angular/common';

import { ModalComponent } from '../../shared';
import type { BudgetResponse } from '../../core/services/planning.service';
import { CategoryService, DEFAULT_INCOME_CATEGORIES } from '../../core/services/category.service';

@Component({
  selector: 'app-income-entry-modal',
  standalone: true,
  imports: [ReactiveFormsModule, ModalComponent, CurrencyPipe],
  templateUrl: './income-entry-modal.component.html',
})
export class IncomeEntryModalComponent implements OnInit {
  readonly submitted = output<{
    description: string;
    value: number;
    category: string;
    account: string;
    date: string | null;
    budgetLimitId?: string;
  }>();
  readonly closed = output<void>();

  readonly accountsInput = input<string[]>([]);
  readonly budgetsInput = input<BudgetResponse[]>([]);

  private readonly fb = inject(FormBuilder);
  private readonly categoryService = inject(CategoryService);

  readonly categories = signal(DEFAULT_INCOME_CATEGORIES.map((c) => c.name));

  form: FormGroup = this.fb.group({
    description: ['', [Validators.required, Validators.maxLength(500)]],
    value: [null as number | null, [Validators.required, Validators.min(0.01)]],
    category: ['', Validators.required],
    account: ['', Validators.required],
    date: [this.todayString()],
    budgetLimitId: [''],
  });

  protected get accounts(): string[] {
    const fromInput = this.accountsInput();
    return fromInput.length > 0 ? fromInput : ['Conta'];
  }

  ngOnInit(): void {
    this.categoryService.getCategories().subscribe({
      next: (cats) => this.categories.set(cats.income.map((c) => c.name)),
    });
  }

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
      budgetLimitId: v.budgetLimitId || undefined,
    });
    this.closed.emit();
  }

  private todayString(): string {
    return new Date().toISOString().slice(0, 10);
  }
}
