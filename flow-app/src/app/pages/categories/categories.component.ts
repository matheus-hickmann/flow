import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CategoryService, CategoryItem, DEFAULT_EXPENSE_CATEGORIES, DEFAULT_INCOME_CATEGORIES } from '../../core/services/category.service';

@Component({
  selector: 'app-categories',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './categories.component.html',
})
export class CategoriesComponent implements OnInit {
  private readonly categoryService = inject(CategoryService);

  readonly activeTab = signal<'expense' | 'income'>('expense');
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal(false);

  expenseCategories = signal<CategoryItem[]>([...DEFAULT_EXPENSE_CATEGORIES]);
  incomeCategories = signal<CategoryItem[]>([...DEFAULT_INCOME_CATEGORIES]);

  readonly newName = signal('');
  readonly newColor = signal('#6b7280');

  ngOnInit(): void {
    this.categoryService.getCategories().subscribe({
      next: (cats) => {
        this.expenseCategories.set([...cats.expense]);
        this.incomeCategories.set([...cats.income]);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      },
    });
  }

  get currentList(): CategoryItem[] {
    return this.activeTab() === 'expense' ? this.expenseCategories() : this.incomeCategories();
  }

  setTab(tab: 'expense' | 'income'): void {
    this.activeTab.set(tab);
    this.newName.set('');
  }

  setNewName(event: Event): void {
    this.newName.set((event.target as HTMLInputElement).value);
  }

  setNewColor(event: Event): void {
    this.newColor.set((event.target as HTMLInputElement).value);
  }

  addCategory(): void {
    const name = this.newName().trim();
    if (!name) return;
    const id = name.toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/\s+/g, '_');
    const item: CategoryItem = { id, name, color: this.newColor() };
    if (this.activeTab() === 'expense') {
      this.expenseCategories.update((list) => [...list, item]);
    } else {
      this.incomeCategories.update((list) => [...list, item]);
    }
    this.newName.set('');
    this.newColor.set('#6b7280');
  }

  removeCategory(id: string): void {
    if (this.activeTab() === 'expense') {
      this.expenseCategories.update((list) => list.filter((c) => c.id !== id));
    } else {
      this.incomeCategories.update((list) => list.filter((c) => c.id !== id));
    }
  }

  updateColor(id: string, event: Event): void {
    const color = (event.target as HTMLInputElement).value;
    const update = (list: CategoryItem[]) =>
      list.map((c) => (c.id === id ? { ...c, color } : c));
    if (this.activeTab() === 'expense') {
      this.expenseCategories.update(update);
    } else {
      this.incomeCategories.update(update);
    }
  }

  save(): void {
    this.saving.set(true);
    this.success.set(false);
    this.error.set(null);
    this.categoryService.saveCategories({
      expense: this.expenseCategories(),
      income: this.incomeCategories(),
    }).subscribe({
      next: () => {
        this.saving.set(false);
        this.success.set(true);
        setTimeout(() => this.success.set(false), 3000);
      },
      error: () => {
        this.saving.set(false);
        this.error.set('Erro ao salvar. Tente novamente.');
      },
    });
  }
}
