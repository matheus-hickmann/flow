import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';

export interface ParsedRow {
  readonly date: string;
  readonly description: string;
  readonly amount: number;
  readonly type: 'DEBIT' | 'CREDIT';
  category: string;
  readonly needsCategory: boolean;
  readonly merchantKey: string;
}

export interface MerchantRule {
  readonly merchantKey: string;
  readonly displayName: string;
  category: string;
}

export interface ImportPreviewResponse {
  readonly rows: ParsedRow[];
  readonly knownRules: MerchantRule[];
}

export interface ImportCommitRequest {
  readonly accountId: string;
  readonly rows: ParsedRow[];
  readonly merchantRules: MerchantRule[];
}

export interface ImportCommitResponse {
  readonly imported: number;
  readonly skipped: number;
}

@Injectable({ providedIn: 'root' })
export class ImportService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private get baseUrl(): string {
    return `${this.env.apiUrl}/api/v1/imports`;
  }

  parseCSV(file: File): Observable<ImportPreviewResponse> {
    const form = new FormData();
    form.append('file', file);
    return this.http.post<ImportPreviewResponse>(`${this.baseUrl}/parse`, form);
  }

  getMerchantRules(): Observable<MerchantRule[]> {
    return this.http.get<MerchantRule[]>(`${this.baseUrl}/merchant-rules`).pipe(
      catchError(() => of([])),
    );
  }

  commit(req: ImportCommitRequest): Observable<ImportCommitResponse> {
    return this.http.post<ImportCommitResponse>(`${this.baseUrl}/commit`, req);
  }
}
