import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';

export interface BudgetResponse {
  readonly id: string;
  readonly category: string;
  readonly limitType: string;
  readonly limitValue: number;
}

@Injectable({ providedIn: 'root' })
export class PlanningService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  /** POST to planning-service; returns 201 on success. */
  submitPlanning(payload: unknown): Observable<boolean> {
    if (!this.env.planningApiUrl) return of(false);
    const url = `${this.env.planningApiUrl}/api/v1/planning/submit`;
    return this.http.post(url, payload, { observe: 'response' }).pipe(
      map((res) => res.status === 201),
      catchError(() => of(false)),
    );
  }

  listBudgets(): Observable<BudgetResponse[]> {
    if (!this.env.planningApiUrl) return of([]);
    return this.http.get<BudgetResponse[]>(`${this.env.planningApiUrl}/api/v1/planning/budgets`).pipe(
      catchError(() => of([])),
    );
  }
}
