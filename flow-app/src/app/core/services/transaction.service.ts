import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';

import { ENVIRONMENT } from '../config';
import type {
  PostTransactionRequestDto,
  TransactionListItemDto,
} from '../models/transaction-api.model';

@Injectable({ providedIn: 'root' })
export class TransactionService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);
  private readonly refreshSubject = new Subject<void>();

  /** Emit after a transaction is created so lists can refetch. */
  readonly refresh$ = this.refreshSubject.asObservable();

  refresh(): void {
    this.refreshSubject.next();
  }

  private get baseUrl(): string {
    return `${this.env.apiUrl}/api/v1/ledger`;
  }

  list(limit = 100, accountId?: string): Observable<TransactionListItemDto[]> {
    if (!this.env.apiUrl) return new Observable((s) => s.next([]));
    let params = new HttpParams().set('limit', String(limit));
    if (accountId) params = params.set('accountId', accountId);
    return this.http.get<TransactionListItemDto[]>(`${this.baseUrl}/transactions`, { params });
  }

  postTransaction(body: PostTransactionRequestDto): Observable<unknown> {
    return this.http.post(`${this.baseUrl}/transactions`, body);
  }
}
