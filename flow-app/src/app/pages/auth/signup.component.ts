import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { switchMap } from 'rxjs/operators';

import { AuthService } from '../../core/services/auth.service';
import { ENVIRONMENT } from '../../core/config';
import { FlowMarkComponent } from '../../shared';

const SUGGESTED_QUESTIONS = [
  'Qual o nome do seu primeiro animal de estimação?',
  'Em qual cidade você nasceu?',
  'Qual o nome da sua escola primária?',
  'Qual o nome do seu melhor amigo de infância?',
  'Qual era o modelo do seu primeiro carro?',
  'Qual o nome de solteira da sua mãe?',
];

function generateUserId(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  let id = '';
  for (let i = 0; i < 12; i++) {
    id += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return id;
}

@Component({
  standalone: true,
  imports: [FormsModule, RouterLink, FlowMarkComponent],
  template: `
    <div class="min-h-screen flex items-center justify-center bg-neutral-50 dark:bg-neutral-950 px-4 py-8">
      <div class="w-full max-w-md space-y-8">
        <div class="text-center">
          <div class="flex justify-center text-neutral-900 dark:text-white">
            <flow-mark variant="wordmark" [height]="48" />
          </div>
          <h1 class="mt-6 text-2xl font-bold text-neutral-900 dark:text-white">Criar conta</h1>
        </div>
        <form (ngSubmit)="onSubmit()" class="space-y-5 rounded-xl bg-white dark:bg-neutral-900 p-6 shadow-sm border border-neutral-200 dark:border-neutral-700">
          @if (errorMessage()) {
            <div class="rounded-lg bg-red-50 dark:bg-red-950 text-red-700 dark:text-red-300 text-sm p-3">{{ errorMessage() }}</div>
          }

          <!-- Nome de exibição -->
          <div>
            <label for="displayName" class="block text-sm font-medium text-neutral-700 dark:text-neutral-300">Como posso te chamar?</label>
            <input id="displayName" type="text" name="displayName" [(ngModel)]="displayName"
                   placeholder="Ex.: João"
                   class="mt-1 block w-full rounded-lg border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-800 px-3 py-2 text-neutral-800 dark:text-neutral-100 shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
            <p class="mt-1 text-xs text-neutral-500 dark:text-neutral-400">Opcional. Exibido no menu do app.</p>
          </div>

          <!-- ID de usuário — pré-gerado, read-only -->
          <div>
            <label for="userId" class="block text-sm font-medium text-neutral-700 dark:text-neutral-300">Seu ID de usuário</label>
            <div class="mt-1 flex gap-2">
              <input id="userId" type="text" name="userId" [ngModel]="userId" readonly
                     class="flex-1 rounded-lg border border-neutral-300 dark:border-neutral-600 bg-neutral-100 dark:bg-neutral-700 px-3 py-2 text-neutral-800 dark:text-neutral-100 font-mono text-sm shadow-sm cursor-default" />
              <button type="button" (click)="regenerateId()"
                      class="shrink-0 rounded-lg border border-neutral-300 dark:border-neutral-600 px-3 py-2 text-sm text-neutral-600 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors"
                      title="Gerar novo ID">
                ↻
              </button>
            </div>
            <div class="mt-2 rounded-lg bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 p-3">
              <p class="text-xs text-amber-800 dark:text-amber-300 font-medium">⚠️ Guarde este ID em um local seguro.</p>
              <p class="text-xs text-amber-700 dark:text-amber-400 mt-1">Junto com sua senha, é a única forma de acessar sua conta. Não é possível recuperá-lo depois.</p>
            </div>
          </div>

          <!-- Senha -->
          <div>
            <label for="password" class="block text-sm font-medium text-neutral-700 dark:text-neutral-300">Senha</label>
            <input id="password" type="password" name="password" [(ngModel)]="password" required autocomplete="new-password"
                   class="mt-1 block w-full rounded-lg border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-800 px-3 py-2 text-neutral-800 dark:text-neutral-100 shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
          </div>

          <!-- Perguntas de recuperação -->
          <div class="border-t border-neutral-200 dark:border-neutral-700 pt-4">
            <p class="text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-1">Perguntas de recuperação</p>
            <p class="text-xs text-neutral-500 dark:text-neutral-400 mb-4">Usadas para verificar sua identidade caso perca o ID ou senha.</p>
            @for (i of questionIndices; track i) {
              <div class="mb-4 space-y-2">
                <label [for]="'q' + i" class="block text-xs font-medium text-neutral-600 dark:text-neutral-400">Pergunta {{ i + 1 }}</label>
                <select [id]="'q' + i" [name]="'question' + i" [(ngModel)]="questions[i]" required
                        class="block w-full rounded-lg border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-800 px-3 py-2 text-sm text-neutral-800 dark:text-neutral-100 shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary">
                  <option value="" disabled>Selecione uma pergunta</option>
                  @for (q of suggestedQuestions; track q) {
                    <option [value]="q">{{ q }}</option>
                  }
                </select>
                <input [id]="'a' + i" type="text" [name]="'answer' + i" [(ngModel)]="answers[i]" required
                       placeholder="Sua resposta"
                       class="block w-full rounded-lg border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-800 px-3 py-2 text-sm text-neutral-800 dark:text-neutral-100 shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary" />
              </div>
            }
          </div>

          <button type="submit" [disabled]="loading()"
                  class="w-full rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 disabled:opacity-50">
            {{ loading() ? 'Criando…' : 'Criar conta' }}
          </button>
        </form>
        <p class="text-center text-sm text-neutral-600 dark:text-neutral-400">
          Já tem conta?
          <a routerLink="/login" class="font-medium text-primary hover:underline">Entrar</a>
        </p>
      </div>
    </div>
  `,
})
export class SignupComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  protected readonly appName = inject(ENVIRONMENT).appName;

  userId = generateUserId();
  displayName = '';
  password = '';
  questions: string[] = ['', '', ''];
  answers: string[] = ['', '', ''];
  protected readonly questionIndices = [0, 1, 2];
  protected readonly suggestedQuestions = SUGGESTED_QUESTIONS;

  protected loading = signal(false);
  protected errorMessage = signal<string | null>(null);

  regenerateId(): void {
    this.userId = generateUserId();
  }

  onSubmit(): void {
    this.errorMessage.set(null);

    for (let i = 0; i < 3; i++) {
      if (!this.questions[i] || !this.answers[i]?.trim()) {
        this.errorMessage.set(`Preencha a pergunta e resposta de recuperação ${i + 1}.`);
        return;
      }
    }
    if (new Set(this.questions).size < 3) {
      this.errorMessage.set('As 3 perguntas de recuperação devem ser diferentes.');
      return;
    }

    this.loading.set(true);
    const recoveryQuestions = this.questions.map((q, i) => ({ question: q, answer: this.answers[i].trim() }));

    this.auth.signup({
      userId: this.userId.trim(),
      password: this.password,
      displayName: this.displayName.trim() || undefined,
    }).pipe(
      switchMap(() => this.auth.saveRecoveryQuestions(recoveryQuestions)),
    ).subscribe({
      next: () => this.router.navigate(['/']),
      error: (err) => {
        this.loading.set(false);
        this.errorMessage.set(err?.error?.message ?? err?.message ?? 'Não foi possível criar a conta.');
      },
      complete: () => this.loading.set(false),
    });
  }
}
