import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, catchError } from 'rxjs';

import { ENVIRONMENT } from '../config';
import type { Account, CreateAccountPayload, AdjustBalancePayload } from '../models/account.model';

@Injectable()
export class AccountService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private get accountsUrl(): string {
    return `${this.env.apiUrl}/api/v1/ledger/accounts`;
  }

  list(): Observable<Account[]> {
    if (!this.env.apiUrl) {
      return of(this.getMockAccounts());
    }
    return this.http.get<Account[]>(this.accountsUrl).pipe(
      catchError(() => of(this.getMockAccounts())),
    );
  }

  create(payload: CreateAccountPayload): Observable<Account> {
    if (!this.env.apiUrl) {
      const mock: Account = {
        id: Date.now(),
        code: payload.name.toUpperCase().replace(/\s+/g, '_').slice(0, 20),
        name: payload.name,
        type: 'ASSET',
        balance: payload.initialBalance,
        color: payload.color || '#3b82f6',
      };
      return of(mock);
    }
    return this.http.post<Account>(this.accountsUrl, payload);
  }

  adjustBalance(id: number, payload: AdjustBalancePayload): Observable<Account> {
    if (!this.env.apiUrl) {
      const accounts = this.getMockAccounts();
      const found = accounts.find((a) => a.id === id);
      const base = found ?? { id, code: '?', name: 'Conta', type: 'ASSET', balance: 0, color: '#3b82f6' };
      return of({ ...base, balance: payload.newBalance });
    }
    return this.http.patch<Account>(`${this.accountsUrl}/${id}/balance`, payload);
  }

  private getMockAccounts(): Account[] {
    return [
      { id: 1, code: 'ITAU', name: 'Itaú', type: 'ASSET', balance: 5200, color: '#0066cc' },
      { id: 2, code: 'SANTANDER', name: 'Santander', type: 'ASSET', balance: 3100, color: '#ec0000' },
      { id: 3, code: 'NUBANK', name: 'Nubank', type: 'ASSET', balance: 890, color: '#820ad1' },
      { id: 4, code: 'DINHEIRO', name: 'Dinheiro', type: 'ASSET', balance: 260, color: '#22c55e' },
    ];
  }
}
