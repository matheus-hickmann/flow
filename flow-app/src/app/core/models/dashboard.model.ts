export interface SummaryCard {
  readonly label: string;
  readonly value: string;
  readonly icon: string;
  readonly bgClass: string;
}

export interface CategorySlice {
  readonly name: string;
  readonly percent: number;
  readonly color: string;
}

export interface BudgetVsActualItem {
  readonly category: string;
  readonly planned: number;
  readonly actual: number;
}

export interface LatestEntry {
  readonly id: string;
  readonly date: string;
  readonly description: string;
  readonly category: string;
  readonly account: string;
  readonly value: string;
  readonly isIncome: boolean;
}
