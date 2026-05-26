import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { map, catchError } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';

export interface ProjectionMonthResult {
  readonly year: number;
  readonly month: number; // 0-based
  readonly label: string;
  readonly projectedIncome: number;
  readonly projectedExpense: number;
  readonly delta: number;
  readonly runningBalance: number;
  readonly isCurrent: boolean;
}

export interface BudgetProjectionResponse {
  readonly currentBalance: number;
  readonly monthlyBudget: number;
  readonly monthlySalary: number;
  readonly hasBudgets: boolean;
  readonly hasSalary: boolean;
  readonly daysRemaining: number;
  readonly months: ProjectionMonthResult[];
}

type RawProjection = {
  currentBalance: number | string;
  monthlyBudget: number | string;
  monthlySalary: number | string;
  hasBudgets: boolean;
  hasSalary: boolean;
  daysRemaining: number;
  months: Array<{
    year: number;
    month: number;
    label: string;
    projectedIncome: number | string;
    projectedExpense: number | string;
    delta: number | string;
    runningBalance: number | string;
    isCurrent: boolean;
  }>;
};

@Injectable({ providedIn: 'root' })
export class ProjectionService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  getBudgetProjection(months = 6): Observable<BudgetProjectionResponse | null> {
    if (!this.env.dashboardApiUrl) return of(null);
    return this.http
      .get<RawProjection>(`${this.env.dashboardApiUrl}/api/v1/dashboard/budget-projection`, {
        params: { months: String(months) },
      })
      .pipe(
        map((r) => ({
          currentBalance: this.toNum(r.currentBalance),
          monthlyBudget: this.toNum(r.monthlyBudget),
          monthlySalary: this.toNum(r.monthlySalary),
          hasBudgets: r.hasBudgets,
          hasSalary: r.hasSalary,
          daysRemaining: r.daysRemaining,
          months: (r.months ?? []).map((m) => ({
            year: m.year,
            month: m.month,
            label: m.label,
            projectedIncome: this.toNum(m.projectedIncome),
            projectedExpense: this.toNum(m.projectedExpense),
            delta: this.toNum(m.delta),
            runningBalance: this.toNum(m.runningBalance),
            isCurrent: m.isCurrent,
          })),
        })),
        catchError(() => of(null)),
      );
  }

  private toNum(val: number | string | undefined): number {
    if (val == null) return 0;
    if (typeof val === 'number') return val;
    return Number(String(val).trim()) || 0;
  }
}
