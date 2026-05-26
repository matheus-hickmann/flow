import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';
import type { Debt, CreateDebtPayload, DebtPaymentPayload } from '../models/debt.model';

@Injectable({ providedIn: 'root' })
export class DebtService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private get baseUrl(): string {
    return `${this.env.apiUrl}/api/v1/debts`;
  }

  list(): Observable<Debt[]> {
    if (!this.env.apiUrl) return of([]);
    return this.http.get<Debt[]>(this.baseUrl).pipe(catchError(() => of([])));
  }

  create(payload: CreateDebtPayload): Observable<{ id: string } | null> {
    if (!this.env.apiUrl) return of(null);
    return this.http.post<{ id: string }>(this.baseUrl, payload).pipe(catchError(() => of(null)));
  }

  recordPayment(id: string, payload: DebtPaymentPayload): Observable<Debt | null> {
    if (!this.env.apiUrl) return of(null);
    return this.http
      .post<Debt>(`${this.baseUrl}/${id}/payment`, payload)
      .pipe(catchError(() => of(null)));
  }

  delete(id: string): Observable<boolean> {
    if (!this.env.apiUrl) return of(false);
    return this.http
      .delete(`${this.baseUrl}/${id}`, { observe: 'response' })
      .pipe(
        map(() => true),
        catchError(() => of(false)),
      );
  }
}
