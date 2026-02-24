import { Component, computed, signal } from '@angular/core';
import { OnInit } from '@angular/core';
import { animate, animation, state, style, transition, trigger } from '@angular/animations';

import { ARIA_LABELS, MONTH_LABELS } from './core/constants/app.constants';
import {
  INITIAL_BUDGET_VS_ACTUAL,
  INITIAL_CATEGORY_SLICES,
  INITIAL_LATEST_ENTRIES,
  INITIAL_SUMMARY_CARDS,
} from './core/constants/dashboard-data.const';
import type {
  BudgetVsActualItem,
  CategorySlice,
  LatestEntry,
  SummaryCard,
} from './core/models/dashboard.model';

const BUDGET_BAR_COLOR = '#1e3a5f';
const ACTUAL_BAR_COLOR = '#86efac';
const BUDGET_CHART_CATEGORIES_PER_PAGE = 6;

const slideMonthAnimation = animation([
  style({ transform: 'translateX({{ dir }}%)', opacity: 0.6 }),
  animate('{{ duration }}ms ease-out', style({ transform: 'translateX(0)', opacity: 1 })),
], { params: { dir: 100, duration: 280 } });

const budgetChartSlideAnimation = animation([
  style({ opacity: 0, transform: 'translateX({{ dir }}%)' }),
  animate('280ms ease-out', style({ opacity: 1, transform: 'translateX(0)' })),
], { params: { dir: 20 } });

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
      transition('false => true', [
        animate('500ms cubic-bezier(0.34, 1.56, 0.64, 1)'),
      ]),
    ]),
    trigger('budgetChartTransition', [
      transition('* => *', [budgetChartSlideAnimation]),
    ]),
  ],
  templateUrl: './app.html',
})
export class App implements OnInit {
  protected readonly budgetBarColor = BUDGET_BAR_COLOR;
  protected readonly actualBarColor = ACTUAL_BAR_COLOR;
  protected readonly ariaLabelMonthPrevious = ARIA_LABELS.MONTH_PREVIOUS;
  protected readonly ariaLabelMonthNext = ARIA_LABELS.MONTH_NEXT;

  private readonly currentDateSignal = signal<{ year: number; month: number }>(App.getInitialDate());
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
  private readonly summaryCardsSignal = signal<readonly SummaryCard[]>(INITIAL_SUMMARY_CARDS);
  private readonly categorySlicesSignal = signal<readonly CategorySlice[]>(INITIAL_CATEGORY_SLICES);
  private readonly budgetVsActualSignal = signal<readonly BudgetVsActualItem[]>(INITIAL_BUDGET_VS_ACTUAL);
  private readonly budgetChartPageSignal = signal(0);
  private readonly budgetChartDirectionSignal = signal(0);
  private readonly latestEntriesSignal = signal<readonly LatestEntry[]>(INITIAL_LATEST_ENTRIES);

  readonly budgetChartPage = this.budgetChartPageSignal.asReadonly();
  protected readonly budgetChartDirection = this.budgetChartDirectionSignal.asReadonly();
  readonly budgetChartTotalPages = computed(() => {
    const total = this.budgetVsActualSignal().length;
    return Math.max(1, Math.ceil(total / BUDGET_CHART_CATEGORIES_PER_PAGE));
  });
  readonly visibleBudgetItems = computed(() => {
    const all = this.budgetVsActualSignal();
    const page = this.budgetChartPageSignal();
    const start = page * BUDGET_CHART_CATEGORIES_PER_PAGE;
    return all.slice(start, start + BUDGET_CHART_CATEGORIES_PER_PAGE);
  });
  readonly budgetChartShowPrev = computed(() => this.budgetChartPageSignal() > 0);
  readonly budgetChartShowNext = computed(() => this.budgetChartPageSignal() < this.budgetChartTotalPages() - 1);

  readonly monthKey = computed(() => {
    const { year, month } = this.currentDateSignal();
    return `${year}-${month}`;
  });
  readonly monthLabel = computed(() => {
    const { year, month } = this.currentDateSignal();
    return `${MONTH_LABELS[month]} ${year}`;
  });
  readonly summaryCards = this.summaryCardsSignal.asReadonly();
  readonly categorySlices = this.categorySlicesSignal.asReadonly();
  readonly budgetVsActual = this.budgetVsActualSignal.asReadonly();
  readonly latestEntries = this.latestEntriesSignal.asReadonly();

  private readonly budgetVsActualMax = computed(() => {
    const items = this.budgetVsActualSignal();
    let max = 0;
    for (const item of items) {
      if (item.budgeted > max) max = item.budgeted;
      if (item.actual > max) max = item.actual;
    }
    return max || 1;
  });

  readonly donutBackground = computed(() => {
    const slices = this.categorySlicesSignal();
    let acc = 0;
    const parts = slices.map((s) => {
      const start = acc;
      acc += s.percent;
      return `${s.color} ${start}% ${acc}%`;
    });
    return `conic-gradient(${parts.join(', ')})`;
  });

  barHeight(value: number): number {
    const max = this.budgetVsActualMax();
    return max > 0 ? Math.round((value / max) * 100) : 0;
  }

  budgetChartPrevPage(): void {
    this.budgetChartDirectionSignal.set(-20);
    this.budgetChartPageSignal.update((p) => Math.max(0, p - 1));
  }

  budgetChartNextPage(): void {
    this.budgetChartDirectionSignal.set(20);
    this.budgetChartPageSignal.update((p) =>
      Math.min(this.budgetChartTotalPages() - 1, p + 1));
  }

  ngOnInit(): void {
    this.scheduleChartFill();
  }

  previousMonth(): void {
    this.chartReadySignal.set(false);
    this.budgetChartPageSignal.set(0);
    if (this.chartFillTimeout != null) {
      clearTimeout(this.chartFillTimeout);
      this.chartFillTimeout = null;
    }
    this.slideDirectionSignal.set(-100);
    this.currentDateSignal.update(({ year, month }) => {
      if (month === 0) return { year: year - 1, month: 11 };
      return { year, month: month - 1 };
    });
    this.scheduleChartFill();
  }

  nextMonth(): void {
    this.chartReadySignal.set(false);
    this.budgetChartPageSignal.set(0);
    if (this.chartFillTimeout != null) {
      clearTimeout(this.chartFillTimeout);
      this.chartFillTimeout = null;
    }
    this.slideDirectionSignal.set(100);
    this.currentDateSignal.update(({ year, month }) => {
      if (month === 11) return { year: year + 1, month: 0 };
      return { year, month: month + 1 };
    });
    this.scheduleChartFill();
  }

  private scheduleChartFill(): void {
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
