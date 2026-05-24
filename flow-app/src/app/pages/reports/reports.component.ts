import { AfterViewInit, Component, ElementRef, OnDestroy, OnInit, ViewChild, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Chart, ChartData, ChartOptions, registerables } from 'chart.js';

import { ReportService, CategorySeries } from '../../core/services/report.service';
import { AccountService } from '../../core/services/account.service';
import { DashboardService } from '../../core/services/dashboard.service';
import type { DailyExpensesResponse } from '../../core/services/dashboard.service';

Chart.register(...registerables);

// Paleta pastel — mapeada por hash de nome de categoria
const FLOW_PASTEL = ['#dcd2ec', '#f4d8c7', '#cce8d6', '#f3e9b9', '#cfdfeb', '#edd1d6', '#cdd9be'];
// Escala do heatmap (clara → mais saturada)
const HEAT_SCALE = ['#ecead8', '#edd1d6', '#e4a9ae', '#c86e75', '#a84439'];

const MONTH_NAMES = ['jan', 'fev', 'mar', 'abr', 'mai', 'jun', 'jul', 'ago', 'set', 'out', 'nov', 'dez'];

function monthLabel(ym: string): string {
  if (!ym) return '';
  const [y, m] = ym.split('-');
  return `${MONTH_NAMES[parseInt(m, 10) - 1]}/${y.slice(2)}`;
}
function monthFull(ym: string): string {
  if (!ym) return '';
  const [y, m] = ym.split('-');
  return `${MONTH_NAMES[parseInt(m, 10) - 1]} ${y}`;
}

function lastNMonthsFirst(n: number): string {
  const d = new Date();
  const m = new Date(d.getFullYear(), d.getMonth() - (n - 1), 1);
  return `${m.getFullYear()}-${String(m.getMonth() + 1).padStart(2, '0')}`;
}

interface CategoryDelta {
  readonly category: string;
  readonly current: number;
  readonly previous: number;
  readonly delta: number;
  readonly barPct: number;
}
interface CrossRow {
  readonly category: string;
  readonly values: readonly number[];
  readonly delta: number;
  readonly spark: string;
}
interface HeatCell {
  readonly day: number;
  readonly value: number;
  readonly color: string;
  readonly fg: string;
  readonly fgSoft: string;
  readonly short: string;
}

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './reports.component.html',
})
export class ReportsComponent implements OnInit, OnDestroy, AfterViewInit {
  @ViewChild('chartCanvas') chartCanvas!: ElementRef<HTMLCanvasElement>;

  // expor Math no template (necessário pra Math.abs nos bindings)
  protected readonly Math = Math;
  protected readonly heatScale = HEAT_SCALE;
  protected readonly dayLabels = ['D', 'S', 'T', 'Q', 'Q', 'S', 'S'];

  private readonly reportService    = inject(ReportService);
  private readonly accountService   = inject(AccountService);
  private readonly dashboardService = inject(DashboardService);
  private chart: Chart | null = null;

  readonly loading = signal(false);
  readonly error   = signal<string | null>(null);

  // Filtros
  fromMonth = lastNMonthsFirst(6);
  toMonth = (() => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
  })();
  type: 'expense' | 'income' | 'all' = 'expense';

  private expenseAccountId = '';
  private incomeAccountId  = '';

  readonly seriesList         = signal<CategorySeries[]>([]);
  readonly selectedCategories = signal<Set<string>>(new Set());
  readonly months             = signal<string[]>([]);
  // Agregação diária do mês corrente p/ heatmap — vem do backend.
  private readonly dailyExpensesSignal = signal<DailyExpensesResponse | null>(null);

  readonly allCategories = computed(() =>
    [...new Set(this.seriesList().map((s) => s.category))],
  );

  readonly filteredSeries = computed(() => {
    const selected = this.selectedCategories();
    return selected.size === 0
      ? this.seriesList()
      : this.seriesList().filter((s) => selected.has(s.category));
  });

  // ─── Labels ───────────────────────────────────────────────────────
  readonly rangeLabel = computed(() => `${monthFull(this.fromMonth)} → ${monthFull(this.toMonth)}`);
  readonly currentMonthLabel = computed(() => {
    const m = this.months();
    return monthLabel(m[m.length - 1] || this.toMonth);
  });
  readonly previousMonthLabel = computed(() => {
    const m = this.months();
    return monthLabel(m[m.length - 2] || '');
  });

  monthShort(ym: string): string {
    return monthLabel(ym);
  }
  numberFmt(v: number): string {
    return Math.round(v).toLocaleString('pt-BR');
  }
  formatBRL(v: number): string {
    const abs = Math.abs(v);
    const s = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL', maximumFractionDigits: 0 }).format(abs);
    return v < 0 ? '−' + s : s;
  }
  formatPct(v: number): string {
    return Math.abs(v).toFixed(1).replace('.', ',') + '%';
  }
  categoryColor(cat: string): string {
    let h = 0;
    for (let i = 0; i < cat.length; i++) h = (h * 31 + cat.charCodeAt(i)) >>> 0;
    return FLOW_PASTEL[h % FLOW_PASTEL.length];
  }

  // ─── Comparação: este mês × anterior × média ──────────────────────
  readonly comparison = computed(() => {
    const series = this.seriesList();
    const months = this.months();
    if (months.length === 0) {
      return { current: 0, previous: 0, avg: 0, delta: 0, vsPrevPct: 0, vsAvgPct: 0, activeCategories: 0 };
    }
    const curIdx = months.length - 1;
    const prevIdx = months.length - 2;
    const totalAt = (idx: number) =>
      idx < 0 ? 0 : series.reduce((s, c) => s + Number(c.byMonth[months[idx]] ?? 0), 0);
    const current = totalAt(curIdx);
    const previous = totalAt(prevIdx);
    const avgWindow = Math.max(1, months.length - 1);
    let avgSum = 0;
    for (let i = 0; i < avgWindow; i++) avgSum += totalAt(i);
    const avg = avgSum / avgWindow;
    const delta = current - previous;
    const vsPrevPct = previous ? ((current - previous) / previous) * 100 : 0;
    const vsAvgPct  = avg ? ((current - avg) / avg) * 100 : 0;
    const activeCategories = series.filter((c) => Number(c.byMonth[months[curIdx]] ?? 0) > 0).length;
    return { current, previous, avg, delta, vsPrevPct, vsAvgPct, activeCategories };
  });

  // ─── Δ por categoria ──────────────────────────────────────────────
  readonly categoryDeltas = computed<CategoryDelta[]>(() => {
    const series = this.seriesList();
    const months = this.months();
    if (months.length < 2) return [];
    const curM = months[months.length - 1];
    const prevM = months[months.length - 2];
    const rows = series.map((s) => {
      const current = Number(s.byMonth[curM] ?? 0);
      const previous = Number(s.byMonth[prevM] ?? 0);
      return { category: s.category, current, previous, delta: current - previous };
    });
    const max = Math.max(1, ...rows.map((r) => Math.abs(r.delta)));
    return rows
      .map((r) => ({ ...r, barPct: (Math.abs(r.delta) / max) * 100 }))
      .sort((a, b) => Math.abs(b.delta) - Math.abs(a.delta));
  });

  // ─── Tabela cruzada categoria × mês ────────────────────────────────
  readonly crossTable = computed<CrossRow[]>(() => {
    const months = this.months();
    if (months.length === 0) return [];
    return this.seriesList().map((s) => {
      const values = months.map((m) => Number(s.byMonth[m] ?? 0));
      const delta = (values[values.length - 1] || 0) - (values[values.length - 2] || 0);
      return { category: s.category, values, delta, spark: this.makeSparkPath(values, 70, 22) };
    });
  });

  readonly totalsRow = computed(() => {
    const months = this.months();
    if (months.length === 0) return null;
    const values = months.map((m) =>
      this.seriesList().reduce((s, c) => s + Number(c.byMonth[m] ?? 0), 0),
    );
    const delta = (values[values.length - 1] || 0) - (values[values.length - 2] || 0);
    return { values, delta, spark: this.makeSparkPath(values, 70, 22) };
  });

  // ─── Heatmap do mês corrente (agregado pelo backend) ───────────────
  readonly heatmapData = computed(() => {
    const d = this.dailyExpensesSignal();
    if (!d) return { dayTotals: [] as number[], year: 0, month: 0 };
    // Backend month is 0-based; legacy code below uses 1-based for date math.
    return { dayTotals: d.dayTotals, year: d.year, month: d.month + 1 };
  });

  readonly heatmapWeeks = computed(() => {
    const { dayTotals, year, month } = this.heatmapData();
    if (dayTotals.length === 0) return [];
    const max = Math.max(...dayTotals);
    const offset = new Date(year, month - 1, 1).getDay();
    const weeks: (HeatCell | null)[][] = [];
    let cur: (HeatCell | null)[] = Array(7).fill(null);
    for (let i = 0; i < offset; i++) cur[i] = null;
    let col = offset;
    for (let d = 1; d <= dayTotals.length; d++) {
      const v = dayTotals[d - 1];
      const t = max ? v / max : 0;
      const idx = t === 0 ? 0 : Math.min(HEAT_SCALE.length - 1, Math.floor(t * (HEAT_SCALE.length - 1) + 0.5));
      const isStrong = idx >= 3;
      const short = v >= 1000 ? (Math.round(v / 100) / 10).toFixed(1) + 'k' : String(Math.round(v));
      cur[col] = {
        day: d,
        value: v,
        color: HEAT_SCALE[idx],
        fg: isStrong ? '#fff' : '#1d1a16',
        fgSoft: isStrong ? 'rgba(255,255,255,0.85)' : '#6b665e',
        short,
      };
      col++;
      if (col === 7) {
        weeks.push(cur);
        cur = Array(7).fill(null);
        col = 0;
      }
    }
    if (col > 0) weeks.push(cur);
    return weeks;
  });

  readonly heatStats = computed(() => {
    const { dayTotals, year, month } = this.heatmapData();
    const total = dayTotals.reduce((s, v) => s + v, 0);
    const active = dayTotals.filter((v) => v > 0).length;
    const empty = dayTotals.filter((v) => v === 0).length;
    let peakIdx = 0;
    for (let i = 0; i < dayTotals.length; i++) if (dayTotals[i] > dayTotals[peakIdx]) peakIdx = i;
    const peakDay = dayTotals.length ? `${peakIdx + 1} ${MONTH_NAMES[month - 1]}` : '—';
    const txCount = this.dailyExpensesSignal()?.transactionCount ?? 0;
    return {
      peakDay,
      peakValue: dayTotals[peakIdx] || 0,
      avgPerActive: active ? total / active : 0,
      activeDays: active,
      totalDays: dayTotals.length,
      emptyDays: empty,
      total,
      txCount,
    };
  });

  // ─── Sparkline path util ──────────────────────────────────────────
  private makeSparkPath(data: readonly number[], w: number, h: number): string {
    if (data.length < 2) return '';
    const min = Math.min(...data), max = Math.max(...data);
    const range = max - min || 1;
    const pad = 2;
    return data
      .map((v, i) => {
        const x = (i / (data.length - 1)) * w;
        const y = h - ((v - min) / range) * (h - pad * 2) - pad;
        return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
      })
      .join(' ');
  }

  // ─── Lifecycle ────────────────────────────────────────────────────
  ngOnInit(): void {
    this.accountService.listAll().subscribe({
      next: (accounts) => {
        this.expenseAccountId = accounts.find((a) => a.name === 'Saída')?.id ?? '';
        this.incomeAccountId  = accounts.find((a) => a.name === 'Entrada')?.id ?? '';
        this.load();
        this.loadHeatmap();
      },
      error: () => {
        this.load();
        this.loadHeatmap();
      },
    });
  }

  ngAfterViewInit(): void {
    // chart é renderizado pelo load()
  }

  ngOnDestroy(): void {
    this.chart?.destroy();
  }

  load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.reportService.getMonthly(
      this.fromMonth, this.toMonth, this.type,
      this.expenseAccountId, this.incomeAccountId,
    ).subscribe({
      next: (data) => {
        this.months.set(data.months);
        this.seriesList.set(data.series);
        this.selectedCategories.set(new Set());
        this.loading.set(false);
        setTimeout(() => this.renderChart(), 50);
        this.loadHeatmap();
      },
      error: () => {
        this.loading.set(false);
        this.error.set('Erro ao carregar dados. Verifique a conexão.');
      },
    });
  }

  /** Heatmap pulls a pre-aggregated daily series from the backend (single call). */
  private loadHeatmap(): void {
    const target = this.toMonth || `${new Date().getFullYear()}-${String(new Date().getMonth() + 1).padStart(2, '0')}`;
    const [y, m] = target.split('-').map(Number);
    this.dashboardService.getDailyExpenses(y, m - 1).subscribe({
      next: (d) => this.dailyExpensesSignal.set(d),
      error: () => this.dailyExpensesSignal.set(null),
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
    const labelColor = '#6b665e';
    const gridColor  = 'rgba(29,26,22,0.06)';

    const datasets = series.map((s) => {
      const color = this.categoryColor(s.category);
      return {
        label: s.category,
        data: months.map((m) => Number(s.byMonth[m] ?? 0)),
        backgroundColor: color + 'cc',
        borderColor: color,
        borderWidth: 1.5,
        borderRadius: 4,
      };
    });

    const data: ChartData = { labels: months.map(monthLabel), datasets };
    const options: ChartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: 'bottom', labels: { color: labelColor, padding: 14, boxWidth: 10, font: { family: 'IBM Plex Sans', size: 11 } } },
        tooltip: { mode: 'index', intersect: false },
      },
      scales: {
        x: { ticks: { color: labelColor, font: { family: 'IBM Plex Mono', size: 10 } }, grid: { color: gridColor } },
        y: { ticks: { color: labelColor, font: { family: 'IBM Plex Mono', size: 10 } }, grid: { color: gridColor }, beginAtZero: true },
      },
    };
    this.chart = new Chart(ctx, { type: 'bar', data, options });
  }
}
