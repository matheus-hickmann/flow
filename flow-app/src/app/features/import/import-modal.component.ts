import {
  Component,
  inject,
  input,
  output,
  signal,
  computed,
  OnInit,
} from '@angular/core';
import { FormsModule } from '@angular/forms';

import { ModalComponent } from '../../shared';
import { CategoryService, DEFAULT_EXPENSE_CATEGORIES, DEFAULT_INCOME_CATEGORIES } from '../../core/services/category.service';
import {
  ImportService,
  type ParsedRow,
  type MerchantRule,
} from '../../core/services/import.service';
import type { Account } from '../../core/models/account.model';

type Step = 'upload' | 'review' | 'done';

@Component({
  selector: 'app-import-modal',
  standalone: true,
  imports: [FormsModule, ModalComponent],
  templateUrl: './import-modal.component.html',
})
export class ImportModalComponent implements OnInit {
  readonly accountsInput = input<Account[]>([]);
  readonly closed = output<void>();
  readonly imported = output<void>();

  private readonly importService = inject(ImportService);
  private readonly categoryService = inject(CategoryService);

  protected readonly step = signal<Step>('upload');
  protected readonly loading = signal(false);
  protected readonly error = signal('');

  // Step 1
  protected selectedFile = signal<File | null>(null);
  protected selectedAccountId = signal('');

  // Step 2
  protected readonly rows = signal<ParsedRow[]>([]);
  protected readonly merchantRules = signal<MerchantRule[]>([]);
  protected readonly allCategories = signal<string[]>(
    DEFAULT_EXPENSE_CATEGORIES.map((c) => c.name),
  );

  // Derived: unique merchants that still need a category
  protected readonly pendingMerchants = computed(() => {
    const seen = new Set<string>();
    return this.rows()
      .filter((r) => r.needsCategory && !seen.has(r.merchantKey) && seen.add(r.merchantKey))
      .map((r) => ({ merchantKey: r.merchantKey, description: r.description }));
  });

  // Category assigned per merchantKey by the user during review
  protected readonly categoryByMerchant = signal<Record<string, string>>({});

  // Preview counts
  protected readonly readyCount = computed(
    () => this.rows().filter((r) => this.resolvedCategory(r) !== '').length,
  );
  protected readonly skipCount = computed(
    () => this.rows().filter((r) => this.resolvedCategory(r) === '').length,
  );

  ngOnInit(): void {
    this.categoryService.getCategories().subscribe({
      next: (cats) => {
        const names = [
          ...cats.expense.map((c) => c.name),
          ...cats.income.map((c) => c.name),
        ];
        this.allCategories.set([...new Set(names)]);
      },
    });
    if (this.accountsInput().length > 0) {
      this.selectedAccountId.set(this.accountsInput()[0].id);
    }
  }

  onFileChange(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.selectedFile.set(input.files?.[0] ?? null);
    this.error.set('');
  }

  onParse(): void {
    const file = this.selectedFile();
    if (!file) return;
    this.loading.set(true);
    this.error.set('');
    this.importService.parseCSV(file).subscribe({
      next: (preview) => {
        this.rows.set(preview.rows);
        this.merchantRules.set(preview.knownRules);
        // Pre-fill the categoryByMerchant map with categories already resolved
        const map: Record<string, string> = {};
        for (const r of preview.rows) {
          if (r.category && !map[r.merchantKey]) {
            map[r.merchantKey] = r.category;
          }
        }
        this.categoryByMerchant.set(map);
        this.loading.set(false);
        this.step.set('review');
      },
      error: (err) => {
        this.error.set(err?.error?.message ?? 'Erro ao processar o arquivo.');
        this.loading.set(false);
      },
    });
  }

  setCategoryForMerchant(merchantKey: string, category: string): void {
    this.categoryByMerchant.update((map) => ({ ...map, [merchantKey]: category }));
  }

  resolvedCategory(row: ParsedRow): string {
    return this.categoryByMerchant()[row.merchantKey] ?? row.category ?? '';
  }

  onCommit(): void {
    const accountId = this.selectedAccountId();
    if (!accountId) return;

    const categoryMap = this.categoryByMerchant();

    // Build resolved rows
    const resolvedRows: ParsedRow[] = this.rows().map((r) => ({
      ...r,
      category: categoryMap[r.merchantKey] ?? r.category ?? '',
    }));

    // Collect new rules (only merchants that needed a category and got one)
    const newRules: MerchantRule[] = this.rows()
      .filter((r) => r.needsCategory && categoryMap[r.merchantKey])
      .reduce<MerchantRule[]>((acc, r) => {
        if (!acc.find((rule) => rule.merchantKey === r.merchantKey)) {
          acc.push({
            merchantKey: r.merchantKey,
            displayName: r.description,
            category: categoryMap[r.merchantKey],
          });
        }
        return acc;
      }, []);

    this.loading.set(true);
    this.importService.commit({ accountId, rows: resolvedRows, merchantRules: newRules }).subscribe({
      next: () => {
        this.loading.set(false);
        this.step.set('done');
        this.imported.emit();
      },
      error: (err) => {
        this.error.set(err?.error?.message ?? 'Erro ao importar transações.');
        this.loading.set(false);
      },
    });
  }

  onClose(): void {
    this.closed.emit();
  }

  protected get accounts(): Account[] {
    return this.accountsInput().filter((a) => !a.isSystem);
  }

  protected formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v);
  }

  protected trackByMerchant(_: number, item: { merchantKey: string }): string {
    return item.merchantKey;
  }

  protected trackByRow(i: number): number {
    return i;
  }
}
