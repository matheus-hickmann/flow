export interface Environment {
  readonly production: boolean;
  readonly appName: string;
  readonly defaultUserName: string;
  /** Ledger API (accounts, transactions). */
  readonly apiUrl: string;
  /** Auth API (login, signup). User-service. */
  readonly authApiUrl: string;
  /** Planning API (budgets, goals). Planning-service. */
  readonly planningApiUrl: string;
  /** Dashboard API (summary, monthly). Dashboard-service. */
  readonly dashboardApiUrl: string;
  /** Account API (recovery questions). Account-service. */
  readonly accountApiUrl: string;
  /** Public-facing origin for absolute OG/social meta URLs (e.g. https://app.example.com). */
  readonly appUrl: string;
}
