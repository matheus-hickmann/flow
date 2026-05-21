import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, Subject, forkJoin } from 'rxjs';
import { map, catchError } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';
import type { TransactionListItemDto } from '../models/transaction-api.model';
import type { BudgetVsActualItem, LatestEntry } from '../models/dashboard.model';

export interface SummaryDto {
  readonly totalBalance: number | string;
  readonly accountsCount: number;
  readonly investmentBalance: number | string;
}

export interface CategoryAmountDto {
  readonly category: string;
  readonly amount: number | string;
}

export interface MonthlySummaryDto {
  readonly monthlyIncome: number | string;
  readonly monthlyExpense: number | string;
  readonly categoryBreakdown: CategoryAmountDto[];
}

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);
  private readonly refreshSubject = new Subject<void>();

  readonly refresh$ = this.refreshSubject.asObservable();

  refresh(): void {
    this.refreshSubject.next();
  }

  getSummary(): Observable<{ totalBalance: number; accountsCount: number; investmentBalance: number }> {
    if (!this.env.dashboardApiUrl) return of({ totalBalance: 0, accountsCount: 0, investmentBalance: 0 });
    return this.http.get<SummaryDto>(`${this.env.dashboardApiUrl}/api/v1/dashboard/summary`).pipe(
      map((s) => ({
        totalBalance: this.toNum(s.totalBalance),
        accountsCount: s.accountsCount,
        investmentBalance: this.toNum(s.investmentBalance),
      })),
      catchError(() => of({ totalBalance: 0, accountsCount: 0, investmentBalance: 0 })),
    );
  }

  /** Monthly income, expense and category breakdown (dashboard cards + donut). */
  getMonthlySummary(year: number, month: number): Observable<MonthlySummaryDto | null> {
    if (!this.env.dashboardApiUrl) return of(null);
    const params = { year: String(year), month: String(month) };
    return this.http.get<MonthlySummaryDto>(`${this.env.dashboardApiUrl}/api/v1/dashboard/summary/monthly`, { params }).pipe(
      catchError(() => of(null)),
    );
  }

  getPlannedVsActual(year: number, month: number): Observable<BudgetVsActualItem[]> {
    if (!this.env.dashboardApiUrl) return of([]);
    const params = { year: String(year), month: String(month) };
    return this.http
      .get<{ category: string; planned: number; actual: number }[]>(
        `${this.env.dashboardApiUrl}/api/v1/dashboard/planned-vs-actual`,
        { params },
      )
      .pipe(
        map((list) =>
          list.map((item) => ({
            category: item.category,
            planned: this.toNum(item.planned),
            actual: this.toNum(item.actual),
          })),
        ),
        catchError(() => of([])),
      );
  }

  /** Latest transactions for dashboard (real data from ledger). */
  getLatestEntries(limit: number): Observable<LatestEntry[]> {
    if (!this.env.apiUrl) return of([]);
    return forkJoin({
      transactions: this.http.get<TransactionListItemDto[]>(
        `${this.env.apiUrl}/api/v1/ledger/transactions`,
        { params: { limit: String(limit) } },
      ),
      accounts: this.http.get<{ id: string; name: string }[]>(
        `${this.env.apiUrl}/api/v1/ledger/accounts`,
        { params: { includeSystem: 'true' } },
      ),
    }).pipe(
      map(({ transactions, accounts }) => {
        const receitasId = accounts.find((a) => a.name === 'Entrada')?.id ?? '';
        const despesasId = accounts.find((a) => a.name === 'Saída')?.id ?? '';
        return transactions.map((tx) => this.toLatestEntry(tx, receitasId, despesasId, accounts));
      }),
      catchError(() => of([])),
    );
  }

  private toLatestEntry(
    tx: TransactionListItemDto,
    receitasId: string,
    despesasId: string,
    accounts: { id: string; name: string }[],
  ): LatestEntry {
    const date = tx.timestamp ? new Date(tx.timestamp) : new Date();
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const userEntry = tx.entries?.find(
      (e) => e.accountId !== receitasId && e.accountId !== despesasId,
    );
    const amount = userEntry ? this.toNum(userEntry.amount) : 0;
    const isIncome = userEntry ? userEntry.type === 'DEBIT' : false;
    const accountName = userEntry
      ? (accounts.find((a) => a.id === userEntry.accountId)?.name ?? '—')
      : '—';
    return {
      id: tx.id,
      date: `${day}/${month}`,
      description: tx.description || '—',
      category: tx.category || 'Outros',
      account: accountName,
      value: (isIncome ? '+ ' : '- ') + new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(Math.abs(amount)),
      isIncome,
    };
  }

  private toNum(val: number | string | undefined): number {
    if (val == null) return 0;
    if (typeof val === 'number') return val;
    const s = String(val).trim().replace(/\./g, '').replace(',', '.');
    return Number(s) || 0;
  }
}
