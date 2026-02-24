import { Component, inject, output } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ModalComponent } from '../../shared';

type PlanningType = 'limit' | 'goal' | 'params';

@Component({
  selector: 'app-planning-entry-modal',
  standalone: true,
  imports: [ReactiveFormsModule, ModalComponent],
  templateUrl: './planning-entry-modal.component.html',
})
export class PlanningEntryModalComponent {
  readonly submitted = output<PlanningSubmitPayload>();
  readonly closed = output<void>();

  private readonly fb = inject(FormBuilder);
  form: FormGroup = this.fb.group({
    planType: ['limit' as PlanningType],
    category: [''],
    limitType: ['ABSOLUTE'],
    limitValue: [null as number | null, [Validators.min(0.01)]],
    goalName: [''],
    expectedReturnRate: [null as number | null, [Validators.min(0)]],
    monthlyContribution: [null as number | null, [Validators.min(0)]],
    targetAmount: [null as number | null, [Validators.min(0)]],
    selicRate: [null as number | null, [Validators.min(0)]],
    ipcaRate: [null as number | null, [Validators.min(0)]],
  });
  protected readonly categories = ['Alimentação', 'Moradia', 'Transporte', 'Saúde', 'Educação', 'Lazer', 'Vestuário', 'Outros'];

  onClose(): void {
    this.closed.emit();
  }

  isSubmitDisabled(): boolean {
    const type = this.form.get('planType')?.value as PlanningType;
    if (type === 'limit') {
      return !this.form.get('category')?.value || !this.form.get('limitValue')?.value || Number(this.form.get('limitValue')?.value) <= 0;
    }
    if (type === 'goal') {
      return !this.form.get('goalName')?.value?.trim();
    }
    if (type === 'params') {
      const selic = this.form.get('selicRate')?.value;
      const ipca = this.form.get('ipcaRate')?.value;
      return selic == null || ipca == null || Number(selic) < 0 || Number(ipca) < 0;
    }
    return true;
  }

  onSubmit(): void {
    const type = this.form.get('planType')?.value as PlanningType;
    const v = this.form.value;
    if (type === 'limit') {
      this.submitted.emit({
        type: 'limit',
        category: v.category,
        limitType: v.limitType,
        limitValue: Number(v.limitValue),
      });
    } else if (type === 'goal') {
      this.submitted.emit({
        type: 'goal',
        name: v.goalName,
        expectedReturnRate: Number(v.expectedReturnRate) || 0,
        monthlyContribution: Number(v.monthlyContribution) || 0,
        targetAmount: v.targetAmount != null && v.targetAmount !== '' ? Number(v.targetAmount) : null,
      });
    } else {
      this.submitted.emit({
        type: 'params',
        selicRate: Number(v.selicRate),
        ipcaRate: Number(v.ipcaRate),
      });
    }
    this.closed.emit();
  }
}

export type PlanningSubmitPayload =
  | { type: 'limit'; category: string; limitType: string; limitValue: number }
  | { type: 'goal'; name: string; expectedReturnRate: number; monthlyContribution: number; targetAmount: number | null }
  | { type: 'params'; selicRate: number; ipcaRate: number };
