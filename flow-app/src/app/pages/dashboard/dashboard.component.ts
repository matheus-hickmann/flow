import { Component, computed, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { Subscription } from 'rxjs';
import { animate, animation, state, style, transition, trigger } from '@angular/animations';

import { ARIA_LABELS, MONTH_LABELS } from '../../core/constants/app.constants';
import {
  INITIAL_BUDGET_VS_ACTUAL,
  INITIAL_CATEGORY_SLICES,
  INITIAL_LATEST_ENTRIES,
  INITIAL_SUMMARY_CARDS,
} from '../../core/constants/dashboard-data.const';
import { AccountService, DashboardService } from '../../core';
import type { ProjectionResponse } from '../../core/services/dashboard.service';
import type { Account } from '../../core/models/account.model';
import type {
  BudgetVsActualItem,
  CategorySlice,
  LatestEntry,
  SummaryCard,
} from '../../core/models/dashboard.model';

// ─── Paleta pastel para mapeamento por categoria/conta ────────────
const FLOW_PASTEL = ['#dcd2ec', '#f4d8c7', '#cce8d6', '#f3e9b9', '#cfdfeb', '#edd1d6', '#cdd9be'];

const slideMonthAnimation = animation([
  style({ transform: 'translateX({{ dir }}%)', opacity: 0.6 }),
  animate('{{ duration }}ms ease-out', style({ transform: 'translateX(0)', opacity: 1 })),
], { params: { dir: 100, duration: 280 } });

@Component({
  selector: 'app-dashboard',
  standalone: true,
  animations: [
    trigger('slideMonth', [
      transition('* => *', [slideMonthAnimation]),
    ]),
    trigger('donutFill', [
      state('false', style({ transform: 'scale(0)', opacity: 0 })),
      state('true', style({ transform: 'scale(1)', opacity: 1 })),
      transition('false => true', [animate('500ms cubic-bezier(0.34, 1.56, 0.64, 1)')]),
    ]),
  ],
  templateUrl: './dashboard.component.html',
})
export class DashboardComponent implements OnInit, OnDestroy {
  protected readonly ariaLabelMonthPrevious = ARIA_LABELS.MONTH_PREVIOUS;
  protected readonly ariaLabelMonthNext = ARIA_LABELS.MONTH_NEXT;

  private readonly dashboardService = inject(DashboardService);
  private readonly accountService = inject(AccountService);

  private readonly currentDateSignal = signal<{ year: number; month: number }>(DashboardComponent.getInitialDate());
  private refreshSub: Subscription | null = null;
  private readonly slideDirectionSignal = signal<number>(0);
  private readonly chartReadySignal = signal(false);
  private chartFillTimeout: ReturnType<typeof setTimeout> | null = null;

  readonly chartReady = this.chartReadySignal.asReadonly();

  readonly slideAnimationState = computed(() => {
    const dir = this.slideDirectionSignal();
    return {
      value: this.monthKey(),
      params: { dir, duration: dir === 0 ? 0 : 280 },
    };
  });

  // ─── State signals ───────────────────────────────────────────────
  private readonly summaryCardsSignal      = signal<readonly SummaryCard[]>(INITIAL_SUMMARY_CARDS);
  private readonly categorySlicesSignal    = signal<readonly CategorySlice[]>(INITIAL_CATEGORY_SLICES);
  private readonly budgetVsActualSignal    = signal<readonly BudgetVsActualItem[]>(INITIAL_BUDGET_VS_ACTUAL);
  private readonly latestEntriesSignal     = signal<readonly LatestEntry[]>(INITIAL_LATEST_ENTRIES);
  private readonly accountsSignal          = signal<Account[]>([]);
  // valores numéricos brutos (decoupled da string formatada em summaryCards)
  private readonly totalBalanceSignal      = signal(0);
  private readonly monthlyIncomeSignal     = signal(0);
  private readonly monthlyExpenseSignal    = signal(0);
  private readonly prevTotalBalanceSignal  = signal(0);
  // série de 30 dias do saldo — populado por DashboardService.getBalanceTrend
  private readonly balanceTrendSignal      = signal<number[]>([]);
  // projeção fim-de-mês — populada por DashboardService.getProjection
  private readonly projectionSignal        = signal<ProjectionResponse | null>(null);

  // ─── Computed readonly exposed to template ───────────────────────
  readonly monthKey  = computed(() => {
    const { year, month } = this.currentDateSignal();
    return `${year}-${month}`;
  });
  readonly monthLabel = computed(() => {
    const { year, month } = this.currentDateSignal();
    return `${MONTH_LABELS[month]} ${year}`;
  });
  readonly summaryCards      = this.summaryCardsSignal.asReadonly();
  readonly categorySlices    = this.categorySlicesSignal.asReadonly();
  readonly budgetVsActual    = this.budgetVsActualSignal.asReadonly();
  readonly latestEntries     = this.latestEntriesSignal.asReadonly();
  readonly accounts          = this.accountsSignal.asReadonly();

  readonly hasCategories  = computed(() => this.categorySlicesSignal().length > 0);
  readonly hasBudgetItems = computed(() => this.budgetVsActualSignal().length > 0);
  readonly hasEntries     = computed(() => this.latestEntriesSignal().length > 0);

  // ─── Saldo total split (parte inteira + decimais) ────────────────
  readonly totalBalanceWhole = computed(() => {
    const v = this.totalBalanceSignal();
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL', maximumFractionDigits: 0 }).format(Math.trunc(v));
  });
  readonly totalBalanceCents = computed(() => {
    const v = this.totalBalanceSignal();
    const cents = Math.abs(v - Math.trunc(v)) * 100;
    return ',' + cents.toFixed(0).padStart(2, '0');
  });

  readonly balanceDeltaPct = computed(() => {
    const cur = this.totalBalanceSignal();
    const prev = this.prevTotalBalanceSignal();
    if (!prev) return '+0,0';
    const d = ((cur - prev) / prev) * 100;
    const sign = d > 0 ? '+' : d < 0 ? '−' : '';
    return sign + Math.abs(d).toFixed(1).replace('.', ',');
  });
  readonly balanceDeltaPositive = computed(() => {
    const cur = this.totalBalanceSignal();
    const prev = this.prevTotalBalanceSignal();
    return cur >= prev;
  });

  // ─── Projeção de fim de mês (dados reais do backend) ─────────────
  readonly projectedBalance = computed(() => {
    const p = this.projectionSignal();
    if (!p) return this.formatBRL(this.totalBalanceSignal());
    return this.formatBRL(p.projectedBalance);
  });
  readonly executionPct = computed(() => {
    const today = new Date().getDate();
    const last = new Date(this.currentDateSignal().year, this.currentDateSignal().month + 1, 0).getDate();
    return Math.round((today / last) * 100);
  });
  readonly projectionHint = computed(() => {
    const p = this.projectionSignal();
    if (!p) return 'aguardando dados…';
    if (p.basis === 'past') return `mês encerrado · faltam 0 dias`;
    if (p.basis === 'future') return `mês futuro · sem projeção`;
    const suffix = p.basis === 'linear-with-salary' ? '· salário previsto' : '· ritmo atual';
    return `faltam ${p.daysRemaining} dias ${suffix}`;
  });

  // ─── Em contas vs gastos ─────────────────────────────────────────
  readonly accountsBalance = computed(() =>
    this.formatBRL(this.accountsSignal().filter(a => a.balance > 0).reduce((s, a) => s + Number(a.balance || 0), 0)),
  );
  readonly accountsNetBalance = computed(() =>
    this.formatBRL(this.accountsSignal().reduce((s, a) => s + Number(a.balance || 0), 0)),
  );
  readonly expensesValue = computed(() => this.formatBRL(this.monthlyExpenseSignal()));
  readonly coverageRatio = computed(() => {
    const pos = this.accountsSignal().filter(a => a.balance > 0).reduce((s, a) => s + Number(a.balance || 0), 0);
    const ex = this.monthlyExpenseSignal();
    if (!ex) return '∞';
    return (pos / ex).toFixed(1).replace('.', ',');
  });

  // ─── Sparkline 30 dias ───────────────────────────────────────────
  private readonly SPARK_W = 260;
  private readonly SPARK_H = 36;
  private readonly SPARK_PAD = 2;

  readonly trendPath = computed(() => this.makeSpark(false));
  readonly trendArea = computed(() => this.makeSpark(true));
  readonly trendStart = computed(() => {
    const t = this.balanceTrendSignal();
    return this.formatBRL(t[0] ?? this.totalBalanceSignal());
  });
  readonly trendDelta = computed(() => {
    const t = this.balanceTrendSignal();
    if (t.length < 2) return '+R$ 0';
    const d = t[t.length - 1] - t[0];
    return (d >= 0 ? '+' : '−') + this.formatBRL(Math.abs(d));
  });

  private makeSpark(area: boolean): string {
    let data = this.balanceTrendSignal();
    if (data.length < 2) {
      const cur = this.totalBalanceSignal() || 1;
      // fallback: gera 30 pontos suaves até o saldo atual
      data = Array.from({ length: 30 }, (_, i) => cur * (0.95 + 0.05 * (i / 29) + 0.01 * Math.sin(i / 3)));
    }
    const min = Math.min(...data), max = Math.max(...data);
    const range = max - min || 1;
    const n = data.length;
    const pts = data.map((v, i) => {
      const x = (i / (n - 1)) * this.SPARK_W;
      const y = this.SPARK_H - ((v - min) / range) * (this.SPARK_H - this.SPARK_PAD * 2) - this.SPARK_PAD;
      return [x, y] as const;
    });
    const path = pts.map((p, i) => (i === 0 ? 'M' : 'L') + p[0].toFixed(1) + ',' + p[1].toFixed(1)).join(' ');
    if (!area) return path;
    return `${path} L ${this.SPARK_W},${this.SPARK_H} L 0,${this.SPARK_H} Z`;
  }

  // ─── Budget computeds ────────────────────────────────────────────
  readonly budgetActualTotal  = computed(() => this.formatBRL(this.budgetVsActualSignal().reduce((s, b) => s + b.actual, 0)));
  readonly budgetPlannedTotal = computed(() => this.formatBRL(this.budgetVsActualSignal().reduce((s, b) => s + b.planned, 0)));
  readonly budgetExecutionPct = computed(() => {
    const items = this.budgetVsActualSignal();
    const p = items.reduce((s, b) => s + b.planned, 0);
    const a = items.reduce((s, b) => s + b.actual, 0);
    return p ? Math.round((a / p) * 100) : 0;
  });
  private readonly budgetMax = computed(() => {
    let m = 0;
    for (const it of this.budgetVsActualSignal()) {
      if (it.planned > m) m = it.planned;
      if (it.actual  > m) m = it.actual;
    }
    return m || 1;
  });
  budgetBarPct(v: number): number {
    return Math.min(100, Math.max(0, (v / this.budgetMax()) * 100));
  }

  // ─── Mapeamento de cor por categoria/conta ───────────────────────
  categoryColor(name: string): string {
    return FLOW_PASTEL[this.hashIdx(name, FLOW_PASTEL.length)];
  }
  accountColor(a: Account): string {
    return FLOW_PASTEL[this.hashIdx(a.id || a.name, FLOW_PASTEL.length)];
  }
  private hashIdx(s: string, n: number): number {
    let h = 0;
    for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0;
    return h % n;
  }

  // ─── Helpers ─────────────────────────────────────────────────────
  formatBRL(v: number, opts: { sign?: boolean } = {}): string {
    const abs = Math.abs(v);
    const s = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(abs);
    if (opts.sign) return (v > 0 ? '+' : v < 0 ? '−' : '') + s;
    return v < 0 ? '−' + s : s;
  }
  deltaLabel(actual: number, planned: number): string {
    const d = actual - planned;
    if (d === 0) return '·';
    return (d > 0 ? '+' : '−') + this.formatBRL(Math.abs(d));
  }

  // ─── Lifecycle ───────────────────────────────────────────────────
  ngOnInit(): void {
    this.scheduleChartFill();
    this.loadAccounts();
    this.loadDashboardData();
    this.refreshSub = this.dashboardService.refresh$.subscribe(() => {
      this.loadAccounts();
      this.loadDashboardData();
    });
  }
  ngOnDestroy(): void {
    this.refreshSub?.unsubscribe();
  }

  private loadAccounts(): void {
    this.accountService.list?.().subscribe?.({
      next: (accs: Account[]) => this.accountsSignal.set(accs ?? []),
      error: () => this.accountsSignal.set([]),
    });
  }

  private loadDashboardData(): void {
    const { year, month } = this.currentDateSignal();
    this.dashboardService.getSummary().subscribe((s) => {
      const total = this.toNum(s.totalBalance);
      this.totalBalanceSignal.set(total);
      const cards = [...this.summaryCardsSignal()];
      cards[0] = { ...cards[0], value: this.formatBRL(total) };
      cards[3] = { ...cards[3], value: this.formatBRL(this.toNum(s.investmentBalance)) };
      this.summaryCardsSignal.set(cards);
    });
    // Sparkline 30d + delta % vs início do mês — dados reais do backend.
    this.dashboardService.getBalanceTrend(30).subscribe((points) => {
      if (points.length === 0) return;
      this.balanceTrendSignal.set(points.map((p) => p.totalBalance));
      // start of month = index (length - today.getDate()), clamped to >= 0
      const today = new Date();
      const idx = Math.max(0, points.length - today.getDate());
      this.prevTotalBalanceSignal.set(points[idx]?.totalBalance ?? points[0]?.totalBalance ?? 0);
    });
    this.dashboardService.getMonthlySummary(year, month).subscribe((m) => {
      if (!m) return;
      const income = this.toNum(m.monthlyIncome);
      const expense = this.toNum(m.monthlyExpense);
      this.monthlyIncomeSignal.set(income);
      this.monthlyExpenseSignal.set(expense);
      const cards = [...this.summaryCardsSignal()];
      cards[1] = { ...cards[1], value: this.formatBRL(income) };
      cards[2] = { ...cards[2], value: this.formatBRL(expense) };
      this.summaryCardsSignal.set(cards);
      const breakdown = m.categoryBreakdown ?? [];
      const total = breakdown.reduce((s, c) => s + this.toNum(c.amount), 0);
      const slices = total > 0
        ? breakdown.map((c, i) => ({
            name: c.category,
            percent: Math.round((this.toNum(c.amount) / total) * 100),
            color: FLOW_PASTEL[i % FLOW_PASTEL.length],
          }))
        : [...INITIAL_CATEGORY_SLICES];
      this.categorySlicesSignal.set(slices);
    });
    this.dashboardService.getLatestEntries(8).subscribe((entries) => {
      this.latestEntriesSignal.set(entries.length > 0 ? entries : [...INITIAL_LATEST_ENTRIES]);
    });
    this.dashboardService.getPlannedVsActual(year, month).subscribe((items) => {
      this.budgetVsActualSignal.set(items);
    });
    this.dashboardService.getProjection(year, month).subscribe((p) => {
      this.projectionSignal.set(p);
    });
  }

  private toNum(val: number | string | undefined | null): number {
    if (val == null) return 0;
    if (typeof val === 'number') return val;
    const s = String(val).trim().replace(/\./g, '').replace(',', '.');
    return Number(s) || 0;
  }

  // ─── Month navigation ────────────────────────────────────────────
  previousMonth(): void {
    this.chartReadySignal.set(false);
    this.slideDirectionSignal.set(-100);
    this.currentDateSignal.update(({ year, month }) => {
      if (month === 0) return { year: year - 1, month: 11 };
      return { year, month: month - 1 };
    });
    this.scheduleChartFill();
    this.loadDashboardData();
  }
  nextMonth(): void {
    this.chartReadySignal.set(false);
    this.slideDirectionSignal.set(100);
    this.currentDateSignal.update(({ year, month }) => {
      if (month === 11) return { year: year + 1, month: 0 };
      return { year, month: month + 1 };
    });
    this.scheduleChartFill();
    this.loadDashboardData();
  }

  private scheduleChartFill(): void {
    if (this.chartFillTimeout != null) {
      clearTimeout(this.chartFillTimeout);
      this.chartFillTimeout = null;
    }
    this.chartFillTimeout = setTimeout(() => {
      this.chartReadySignal.set(true);
      this.chartFillTimeout = null;
    }, 320);
  }

  private static getInitialDate(): { year: number; month: number } {
    const now = new Date();
    return { year: now.getFullYear(), month: now.getMonth() };
  }
}
