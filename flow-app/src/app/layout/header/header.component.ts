import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';

import { ROUTES } from '../../core/constants/app.constants';
import { ENVIRONMENT } from '../../core/config';
import { AuthService } from '../../core/services/auth.service';
import { ThemeService } from '../../core/services/theme.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './header.component.html',
})
export class HeaderComponent {
  protected readonly routes = ROUTES;
  protected readonly appName = inject(ENVIRONMENT).appName;
  protected readonly auth = inject(AuthService);
  protected readonly theme = inject(ThemeService);
  private readonly router = inject(Router);

  protected readonly isMobileOpen = signal(false);

  protected get displayUser(): string {
    const u = this.auth.user();
    if (!u) return inject(ENVIRONMENT).defaultUserName;
    return u.displayName ?? u.name ?? u.userId;
  }

  toggleMobile(): void {
    this.isMobileOpen.update((v) => !v);
  }

  closeMobile(): void {
    this.isMobileOpen.set(false);
  }

  logout(): void {
    this.auth.logout();
    this.router.navigate(['/login']);
  }
}
