import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { ENVIRONMENT } from '../config';

export interface CategoryItem {
  readonly id: string;
  readonly name: string;
  readonly color: string;
}

export interface CategoryList {
  readonly expense: CategoryItem[];
  readonly income: CategoryItem[];
}

export const DEFAULT_EXPENSE_CATEGORIES: CategoryItem[] = [
  { id: 'alimentacao', name: 'Alimentação', color: '#f97316' },
  { id: 'moradia', name: 'Moradia', color: '#8b5cf6' },
  { id: 'transporte', name: 'Transporte', color: '#3b82f6' },
  { id: 'saude', name: 'Saúde', color: '#ef4444' },
  { id: 'educacao', name: 'Educação', color: '#eab308' },
  { id: 'lazer', name: 'Lazer', color: '#ec4899' },
  { id: 'vestuario', name: 'Vestuário', color: '#14b8a6' },
  { id: 'outros', name: 'Outros', color: '#6b7280' },
];

export const DEFAULT_INCOME_CATEGORIES: CategoryItem[] = [
  { id: 'salario', name: 'Salário', color: '#22c55e' },
  { id: 'freelance', name: 'Freelance', color: '#10b981' },
  { id: 'investimentos', name: 'Investimentos', color: '#f59e0b' },
  { id: 'aluguel', name: 'Aluguel recebido', color: '#06b6d4' },
  { id: 'outros', name: 'Outros', color: '#6b7280' },
];

@Injectable({ providedIn: 'root' })
export class CategoryService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private get url(): string {
    return `${this.env.apiUrl}/api/v1/categories`;
  }

  getCategories(): Observable<CategoryList> {
    if (!this.env.apiUrl) {
      return of({ expense: DEFAULT_EXPENSE_CATEGORIES, income: DEFAULT_INCOME_CATEGORIES });
    }
    return this.http.get<CategoryList>(this.url).pipe(
      catchError(() => of({ expense: DEFAULT_EXPENSE_CATEGORIES, income: DEFAULT_INCOME_CATEGORIES })),
    );
  }

  saveCategories(payload: CategoryList): Observable<CategoryList> {
    return this.http.put<CategoryList>(this.url, payload);
  }
}
