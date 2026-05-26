import { Component, computed, inject, OnInit, signal } from '@angular/core';

import { ProjectionService, BudgetProjectionResponse } from '../../core/services/projection.service';
import { DashboardService } from '../../core/services/dashboard.service';
import type { BudgetVsActualItem } from '../../core/models/dashboard.model';
import { CurrencyBrlPipe } from '../../shared';

@Component({
  selector: 'app-projections',
  standalone: true,
  imports: [CurrencyBrlPipe],
  templateUrl: './projections.component.html',
})
export class ProjectionsComponent implements OnInit {
  private readonly projectionService = inject(ProjectionService);
  private readonly dashboardService = inject(DashboardService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly data = signal<BudgetProjectionResponse | null>(null);
  readonly plannedVsActual = signal<BudgetVsActualItem[]>([]);

  readonly projectedEOM = computed(() => {
    const d = this.data();
    if (!d || d.months.length === 0) return null;
    return d.months[0];
  });

  readonly projectedFinal = computed(() => {
    const d = this.data();
    if (!d || d.months.length === 0) return null;
    return d.months[d.months.length - 1];
  });

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.error.set(null);

    const now = new Date();
    const year = now.getFullYear();
    const month = now.getMonth(); // 0-based

    this.projectionService.getBudgetProjection(6).subscribe({
      next: (result) => {
        this.data.set(result);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Erro ao carregar projeções.');
        this.loading.set(false);
      },
    });

    this.dashboardService.getPlannedVsActual(year, month).subscribe({
      next: (items) => this.plannedVsActual.set(items.filter((i) => i.planned > 0)),
    });
  }

  budgetPct(item: BudgetVsActualItem): number {
    if (item.planned === 0) return 0;
    return Math.min(100, Math.round((item.actual / item.planned) * 100));
  }

  isOverBudget(item: BudgetVsActualItem): boolean {
    return item.planned > 0 && item.actual > item.planned;
  }

  overAmount(item: BudgetVsActualItem): number {
    return item.actual - item.planned;
  }
}
