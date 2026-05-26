import { Injectable, signal, effect, computed } from '@angular/core';

export type FlowTheme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'flow.theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  readonly preference = signal<FlowTheme>(this.readInitial());
  readonly resolved = signal<'light' | 'dark'>(this.resolveInitial());

  /** Compatibilidade com código que usa `theme.isDark()`. */
  readonly isDark = computed(() => this.resolved() === 'dark');

  private mql = typeof window !== 'undefined'
    ? window.matchMedia('(prefers-color-scheme: dark)')
    : null;

  constructor() {
    this.mql?.addEventListener('change', () => {
      if (this.preference() === 'system') {
        this.resolved.set(this.mql!.matches ? 'dark' : 'light');
      }
    });
    effect(() => {
      const pref = this.preference();
      const next: 'light' | 'dark' =
        pref === 'system' ? (this.mql?.matches ? 'dark' : 'light') : pref;
      this.resolved.set(next);
      this.apply(next);
      try { localStorage.setItem(STORAGE_KEY, pref); } catch {}
    });
  }

  set(theme: FlowTheme): void {
    this.preference.set(theme);
  }

  toggle(): void {
    this.preference.set(this.resolved() === 'dark' ? 'light' : 'dark');
  }

  private readInitial(): FlowTheme {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
        ?? localStorage.getItem('flow_theme');
      if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
      if (stored === 'true') return 'dark';
      if (stored === 'false') return 'light';
    } catch {}
    return 'system';
  }

  private resolveInitial(): 'light' | 'dark' {
    const pref = this.readInitial();
    if (pref === 'light' || pref === 'dark') return pref;
    return this.mql?.matches ? 'dark' : 'light';
  }

  private apply(theme: 'light' | 'dark'): void {
    if (typeof document === 'undefined') return;
    document.body.dataset['theme'] = theme;
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }
}
