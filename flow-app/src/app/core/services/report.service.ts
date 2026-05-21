import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';

export interface CategorySeries {
  readonly category: string;
  readonly color: string | null;
  readonly type: 'expense' | 'income';
  readonly byMonth: Record<string, number>;
}

export interface MonthlyReport {
  readonly months: string[];
  readonly series: CategorySeries[];
}

@Injectable({ providedIn: 'root' })
export class ReportService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  getMonthly(
    from: string,
    to: string,
    type: 'expense' | 'income' | 'all',
    expenseAccountId?: string,
    incomeAccountId?: string,
  ): Observable<MonthlyReport> {
    if (!this.env.apiUrl) {
      return of({ months: [], series: [] });
    }
    let params = new HttpParams().set('from', from).set('to', to).set('type', type);
    if (expenseAccountId) params = params.set('expenseAccountId', expenseAccountId);
    if (incomeAccountId) params = params.set('incomeAccountId', incomeAccountId);
    return this.http.get<MonthlyReport>(`${this.env.apiUrl}/api/v1/reports/monthly`, { params }).pipe(
      catchError(() => of({ months: [], series: [] })),
    );
  }
}
