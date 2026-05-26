import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';

import { AuthService } from '../../core/services/auth.service';
import { ENVIRONMENT } from '../../core/config';
import { FlowMarkComponent } from '../../shared';

@Component({
  standalone: true,
  imports: [FormsModule, RouterLink, FlowMarkComponent],
  template: `
    <div class="min-h-screen flex items-center justify-center bg-neutral-50 px-4">
      <div class="w-full max-w-sm space-y-8">
        <div class="text-center">
          <div class="flex justify-center text-neutral-900">
            <flow-mark variant="wordmark" [height]="48" />
          </div>
          <h1 class="mt-6 text-2xl font-bold text-neutral-900">Entrar</h1>
        </div>
        <form (ngSubmit)="onSubmit()" class="space-y-6 rounded-xl bg-white p-6 shadow-sm border border-neutral-200">
          @if (errorMessage()) {
            <div class="rounded-lg bg-red-50 text-red-700 text-sm p-3">{{ errorMessage() }}</div>
          }
          <div>
            <label for="userId" class="block text-sm font-medium text-neutral-700">ID de usuário</label>
            <input
              id="userId"
              type="text"
              name="userId"
              [(ngModel)]="userId"
              required
              autocomplete="username"
              class="mt-1 block w-full rounded-lg border border-neutral-300 px-3 py-2 shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
          <div>
            <label for="password" class="block text-sm font-medium text-neutral-700">Senha</label>
            <input
              id="password"
              type="password"
              name="password"
              [(ngModel)]="password"
              required
              autocomplete="current-password"
              class="mt-1 block w-full rounded-lg border border-neutral-300 px-3 py-2 shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
          <button
            type="submit"
            [disabled]="loading()"
            class="w-full rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 disabled:opacity-50"
          >
            {{ loading() ? 'Entrando…' : 'Entrar' }}
          </button>
        </form>
        <p class="text-center text-sm text-neutral-600">
          Não tem conta?
          <a routerLink="/criar-conta" class="font-medium text-primary hover:underline">Criar conta</a>
        </p>
      </div>
    </div>
  `,
})
export class LoginComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  protected readonly appName = inject(ENVIRONMENT).appName;

  userId = '';
  password = '';
  protected loading = signal(false);
  protected errorMessage = signal<string | null>(null);

  onSubmit(): void {
    this.errorMessage.set(null);
    this.loading.set(true);
    this.auth.login({ userId: this.userId.trim(), password: this.password }).subscribe({
      next: () => {
        const redirect = this.route.snapshot.queryParamMap.get('redirect');
        this.router.navigateByUrl(redirect ?? '/');
      },
      error: (err) => {
        this.loading.set(false);
        const msg = err?.error?.message ?? err?.message ?? 'ID de usuário ou senha inválidos.';
        this.errorMessage.set(msg);
      },
      complete: () => this.loading.set(false),
    });
  }
}
