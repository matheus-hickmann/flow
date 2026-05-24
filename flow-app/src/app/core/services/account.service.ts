import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, catchError, map } from 'rxjs';

import { ENVIRONMENT } from '../config';
import type { Account, CreateAccountPayload, AdjustBalancePayload, RenameAccountPayload, UpdateAccountPayload } from '../models/account.model';

function parseBalance(val: number | string | undefined): number {
  if (val == null) return 0;
  if (typeof val === 'number') return val;
  const s = String(val).trim().replace(/\./g, '').replace(',', '.');
  return Number(s) || 0;
}

@Injectable({ providedIn: 'root' })
export class AccountService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private get accountsUrl(): string {
    return `${this.env.apiUrl}/api/v1/ledger/accounts`;
  }

  list(): Observable<Account[]> {
    if (!this.env.apiUrl) {
      return of([]);
    }
    return this.http.get<Account[]>(this.accountsUrl).pipe(
      map((list) => list.map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) }))),
      catchError(() => of([])),
    );
  }

  listAll(): Observable<Account[]> {
    return this.http.get<Account[]>(`${this.accountsUrl}?includeSystem=true`).pipe(
      map((list) => list.map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) }))),
      catchError(() => of([])),
    );
  }

  create(payload: CreateAccountPayload): Observable<Account> {
    if (!this.env.apiUrl) {
      const mock: Account = {
        id: `mock-${Date.now()}`,
        code: payload.name.toUpperCase().replace(/\s+/g, '_').slice(0, 20),
        name: payload.name,
        type: 'ASSET',
        balance: payload.initialBalance,
        color: payload.color || '#3b82f6',
      };
      return of(mock);
    }
    return this.http.post<Account>(this.accountsUrl, payload).pipe(
      map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) })),
    );
  }

  rename(id: string, payload: RenameAccountPayload): Observable<Account> {
    if (!this.env.apiUrl) {
      return of({ id, code: id, name: payload.name, type: 'ASSET', balance: 0, color: '#3b82f6' });
    }
    return this.http.patch<Account>(`${this.accountsUrl}/${id}`, payload).pipe(
      map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) })),
    );
  }

  update(id: string, payload: UpdateAccountPayload): Observable<Account> {
    return this.http.patch<Account>(`${this.accountsUrl}/${id}`, payload).pipe(
      map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) })),
    );
  }

  delete(id: string): Observable<void> {
    if (!this.env.apiUrl) {
      return of(undefined);
    }
    return this.http.delete<void>(`${this.accountsUrl}/${id}`);
  }

  adjustBalance(id: string, payload: AdjustBalancePayload): Observable<Account> {
    if (!this.env.apiUrl) {
      const accounts = this.getMockAccounts();
      const found = accounts.find((a) => a.id === id);
      const base = found ?? { id, code: '?', name: 'Conta', type: 'ASSET', balance: 0, color: '#3b82f6' };
      return of({ ...base, balance: payload.newBalance });
    }
    return this.http.patch<Account>(`${this.accountsUrl}/${id}/balance`, payload).pipe(
      map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) })),
    );
  }

  private getMockAccounts(): Account[] {
    return [
      { id: '1', code: 'ITAU', name: 'Itaú', type: 'ASSET', balance: 5200, color: '#0066cc' },
      { id: '2', code: 'SANTANDER', name: 'Santander', type: 'ASSET', balance: 3100, color: '#ec0000' },
      { id: '3', code: 'NUBANK', name: 'Nubank', type: 'ASSET', balance: 890, color: '#820ad1' },
      { id: '4', code: 'DINHEIRO', name: 'Dinheiro', type: 'ASSET', balance: 260, color: '#22c55e' },
    ];
  }
}
