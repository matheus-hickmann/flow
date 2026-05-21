import { Injectable, signal } from '@angular/core';

const STORAGE_KEY = 'flow_theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly darkSignal = signal<boolean>(this.loadPreference());
  readonly isDark = this.darkSignal.asReadonly();

  constructor() {
    this.applyTheme(this.darkSignal());
  }

  toggle(): void {
    this.set(!this.darkSignal());
  }

  set(dark: boolean): void {
    this.darkSignal.set(dark);
    this.applyTheme(dark);
    localStorage.setItem(STORAGE_KEY, dark ? 'dark' : 'light');
  }

  private applyTheme(dark: boolean): void {
    if (typeof document !== 'undefined') {
      document.documentElement.classList.toggle('dark', dark);
    }
  }

  private loadPreference(): boolean {
    if (typeof localStorage === 'undefined') return false;
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) return stored === 'dark';
    return typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches;
  }
}
