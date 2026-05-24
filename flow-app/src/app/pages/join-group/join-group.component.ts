import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { NgClass } from '@angular/common';

import { FamilyService } from '../../core/services/family.service';
import { AuthService } from '../../core/services/auth.service';
import type { InvitePreview } from '../../core/models/family.model';

@Component({
  selector: 'app-join-group',
  standalone: true,
  imports: [NgClass],
  templateUrl: './join-group.component.html',
})
export class JoinGroupComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly familyService = inject(FamilyService);
  private readonly authService = inject(AuthService);

  readonly token = signal('');
  readonly preview = signal<InvitePreview | null>(null);
  readonly loading = signal(true);
  readonly accepting = signal(false);
  readonly error = signal<string | null>(null);
  readonly done = signal(false);

  readonly isLoggedIn = this.authService.user;

  ngOnInit(): void {
    const t = this.route.snapshot.paramMap.get('token') ?? '';
    this.token.set(t);
    this.familyService.getInvitePreview(t).subscribe({
      next: (p) => {
        this.preview.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Não foi possível carregar o convite.');
        this.loading.set(false);
      },
    });
  }

  accept(): void {
    if (!this.isLoggedIn()) {
      this.router.navigate(['/login'], { queryParams: { redirect: `/entrar/${this.token()}` } });
      return;
    }
    this.accepting.set(true);
    this.familyService.acceptInvite(this.token()).subscribe({
      next: () => {
        this.accepting.set(false);
        this.done.set(true);
      },
      error: (err) => {
        this.accepting.set(false);
        const msg = err?.error?.message ?? err?.message ?? 'Erro ao aceitar convite.';
        this.error.set(msg);
      },
    });
  }

  goToFamily(): void {
    this.router.navigate(['/familia']);
  }
}
