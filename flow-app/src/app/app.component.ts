import { Component, inject, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { ARIA_LABELS, FAB_OPTIONS } from './core/constants/app.constants';
import { HeaderComponent, FabMenuComponent, type FabOptionId } from './layout';
import {
  ExpenseEntryModalComponent,
  IncomeEntryModalComponent,
  PlanningEntryModalComponent,
  type PlanningSubmitPayload,
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
  ],
  templateUrl: './app.component.html',
})
export class AppComponent {
  protected readonly ariaLabelFabToggle = ARIA_LABELS.FAB_TOGGLE;
  protected readonly fabOptions = FAB_OPTIONS;

  private readonly isFabOpenSignal = signal(false);
  readonly isFabOpen = this.isFabOpenSignal.asReadonly();

  toggleFab(): void {
    this.isFabOpenSignal.update((open) => !open);
  }

  private readonly openModalSignal = signal<'expense' | 'income' | 'planning' | null>(null);
  readonly openModal = this.openModalSignal.asReadonly();

  onFabOptionSelect(optionId: FabOptionId): void {
    this.openModalSignal.set(optionId);
    this.isFabOpenSignal.set(false);
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
  }): void {
    // TODO: enviar para API (ledger-service) e atualizar lista
    console.log('Despesa cadastrada:', payload);
  }

  onIncomeSubmitted(payload: {
    description: string;
    value: number;
    category: string;
    account: string;
    date: string | null;
  }): void {
    // TODO: enviar para API (ledger-service) e atualizar lista
    console.log('Receita cadastrada:', payload);
  }

  onPlanningSubmitted(payload: PlanningSubmitPayload): void {
    // TODO: enviar para API (plan-service) conforme tipo
    console.log('Planejamento cadastrado:', payload);
  }
}
