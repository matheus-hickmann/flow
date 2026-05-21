import { Component, inject, OnInit, OnDestroy, signal, computed, ElementRef, ViewChild, AfterViewInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Chart, ChartData, ChartOptions, registerables } from 'chart.js';

import { ReportService, CategorySeries } from '../../core/services/report.service';
import { AccountService } from '../../core/services/account.service';

Chart.register(...registerables);

const PALETTE = [
  '#22c55e', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6',
  '#ec4899', '#14b8a6', '#f97316', '#06b6d4', '#eab308',
];

function monthLabel(ym: string): string {
  const [y, m] = ym.split('-');
  const months = ['Jan', 'Fev', 'Mar', 'Abr', 'Mai', 'Jun', 'Jul', 'Ago', 'Set', 'Out', 'Nov', 'Dez'];
  return `${months[parseInt(m, 10) - 1]}/${y.slice(2)}`;
}

function lastNMonths(n: number): string {
  const result: string[] = [];
  const d = new Date();
  for (let i = n - 1; i >= 0; i--) {
    const m = new Date(d.getFullYear(), d.getMonth() - i, 1);
    result.push(`${m.getFullYear()}-${String(m.getMonth() + 1).padStart(2, '0')}`);
  }
  return result[0];
}

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './reports.component.html',
})
export class ReportsComponent implements OnInit, OnDestroy {
  @ViewChild('chartCanvas') chartCanvas!: ElementRef<HTMLCanvasElement>;

  private readonly reportService = inject(ReportService);
  private readonly accountService = inject(AccountService);
  private chart: Chart | null = null;

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);

  // Filters
  fromMonth = lastNMonths(6);
  toMonth = (() => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
  })();
  type: 'expense' | 'income' | 'all' = 'expense';

  private expenseAccountId = '';
  private incomeAccountId = '';

  readonly seriesList = signal<CategorySeries[]>([]);
  readonly selectedCategories = signal<Set<string>>(new Set());
  readonly months = signal<string[]>([]);

  readonly allCategories = computed(() =>
    [...new Set(this.seriesList().map((s) => s.category))]
  );

  readonly filteredSeries = computed(() => {
    const selected = this.selectedCategories();
    return selected.size === 0
      ? this.seriesList()
      : this.seriesList().filter((s) => selected.has(s.category));
  });

  ngOnInit(): void {
    this.accountService.listAll().subscribe({
      next: (accounts) => {
        this.expenseAccountId = accounts.find((a) => a.name === 'Saída')?.id ?? '';
        this.incomeAccountId = accounts.find((a) => a.name === 'Entrada')?.id ?? '';
        this.load();
      },
      error: () => this.load(),
    });
  }

  ngOnDestroy(): void {
    this.chart?.destroy();
  }

  load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.reportService.getMonthly(
      this.fromMonth,
      this.toMonth,
      this.type,
      this.expenseAccountId,
      this.incomeAccountId,
    ).subscribe({
      next: (data) => {
        this.months.set(data.months);
        this.seriesList.set(data.series);
        this.selectedCategories.set(new Set());
        this.loading.set(false);
        setTimeout(() => this.renderChart(), 50);
      },
      error: () => {
        this.loading.set(false);
        this.error.set('Erro ao carregar dados. Verifique a conexão.');
      },
    });
  }

  toggleCategory(cat: string): void {
    this.selectedCategories.update((set) => {
      const next = new Set(set);
      if (next.has(cat)) next.delete(cat);
      else next.add(cat);
      return next;
    });
    setTimeout(() => this.renderChart(), 10);
  }

  renderChart(): void {
    if (!this.chartCanvas) return;
    const ctx = this.chartCanvas.nativeElement.getContext('2d');
    if (!ctx) return;

    this.chart?.destroy();

    const months = this.months();
    const series = this.filteredSeries();
    const isDark = document.documentElement.classList.contains('dark');
    const gridColor = isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)';
    const labelColor = isDark ? '#cbd5e1' : '#475569';

    const datasets = series.map((s, i) => ({
      label: s.category,
      data: months.map((m) => Number(s.byMonth[m] ?? 0)),
      backgroundColor: PALETTE[i % PALETTE.length] + '99',
      borderColor: PALETTE[i % PALETTE.length],
      borderWidth: 2,
      borderRadius: 4,
    }));

    const data: ChartData = {
      labels: months.map(monthLabel),
      datasets,
    };

    const options: ChartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: 'bottom', labels: { color: labelColor, padding: 16, boxWidth: 12 } },
        tooltip: { mode: 'index', intersect: false },
      },
      scales: {
        x: { stacked: false, ticks: { color: labelColor }, grid: { color: gridColor } },
        y: { stacked: false, ticks: { color: labelColor }, grid: { color: gridColor },
          beginAtZero: true },
      },
    };

    this.chart = new Chart(ctx, { type: 'bar', data, options });
  }
}
