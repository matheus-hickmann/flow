import { computed, inject, Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { EMPTY, Observable, tap } from 'rxjs';

import { ENVIRONMENT } from '../config';

const STORAGE_TOKEN = 'flow_token';
const STORAGE_USER = 'flow_user';

/** No personal data: only userId for multi-tenancy. Optional name from backend /me. */
export interface AuthUser {
  userId: string;
  name?: string;
  displayName?: string;
}

/** Response from GET /api/v1/users/me */
export interface MeResponse {
  userId: string;
  source?: string;
  name?: string;
  displayName?: string;
}

export interface LoginRequest {
  userId: string;
  password: string;
}

export interface SignupRequest {
  userId: string;
  password: string;
  displayName?: string;
}

export interface AuthResponse {
  accessToken: string;
  userId: string;
  displayName?: string;
}

export interface RecoveryQuestion {
  question: string;
  answer: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private readonly tokenSignal = signal<string | null>(this.loadStoredToken());
  private readonly userSignal = signal<AuthUser | null>(this.loadStoredUser());

  readonly token = this.tokenSignal.asReadonly();
  readonly user = this.userSignal.asReadonly();
  readonly isLoggedIn = computed(() => !!this.tokenSignal());

  /** True when auth is configured (authApiUrl set); then login is required. */
  useAuth(): boolean {
    return !!this.env.authApiUrl;
  }

  login(body: LoginRequest): Observable<AuthResponse> {
    const url = `${this.env.authApiUrl}/api/v1/auth/login`;
    return this.http.post<AuthResponse>(url, body).pipe(
      tap((res) => {
        this.setSession(res);
        this.refreshMe().subscribe();
      }),
    );
  }

  signup(body: SignupRequest): Observable<AuthResponse> {
    const url = `${this.env.authApiUrl}/api/v1/auth/signup`;
    return this.http.post<AuthResponse>(url, body).pipe(
      tap((res) => {
        this.setSession(res);
        this.refreshMe().subscribe();
      }),
    );
  }

  /** Busca os dados da conta logada no backend (GET /api/v1/users/me). */
  getMe(): Observable<MeResponse> {
    if (!this.env.authApiUrl || !this.tokenSignal()) {
      return EMPTY;
    }
    const url = `${this.env.authApiUrl}/api/v1/users/me`;
    return this.http.get<MeResponse>(url);
  }

  /** Busca /me e atualiza user signal e sessionStorage. Chamado após login/signup e no init quando logado. */
  refreshMe(): Observable<MeResponse | null> {
    return this.getMe().pipe(
      tap((me) => {
        const user: AuthUser = {
          userId: me.userId,
          ...(me.name && { name: me.name }),
          ...(me.displayName && { displayName: me.displayName }),
        };
        this.userSignal.set(user);
        if (typeof sessionStorage !== 'undefined') {
          sessionStorage.setItem(STORAGE_USER, JSON.stringify(user));
        }
      }),
    );
  }

  /** Chama refreshMe() se estiver logado e auth configurado. Usar no init do app. */
  refreshUserFromBackend(): void {
    if (this.useAuth() && this.tokenSignal()) {
      this.refreshMe().subscribe();
    }
  }

  /** Salva as 3 perguntas de recuperação no account-service (chamado após signup). */
  saveRecoveryQuestions(questions: RecoveryQuestion[]): Observable<void> {
    if (!this.env.authApiUrl) return EMPTY;
    const url = `${this.env.authApiUrl}/api/v1/users/me/recovery-questions`;
    return this.http.post<void>(url, { questions });
  }

  logout(): void {
    sessionStorage.removeItem(STORAGE_TOKEN);
    sessionStorage.removeItem(STORAGE_USER);
    this.tokenSignal.set(null);
    this.userSignal.set(null);
  }

  private setSession(res: AuthResponse): void {
    const user: AuthUser = { userId: res.userId, ...(res.displayName && { displayName: res.displayName }) };
    sessionStorage.setItem(STORAGE_TOKEN, res.accessToken);
    sessionStorage.setItem(STORAGE_USER, JSON.stringify(user));
    this.tokenSignal.set(res.accessToken);
    this.userSignal.set(user);
  }

  private loadStoredToken(): string | null {
    if (typeof sessionStorage === 'undefined') return null;
    return sessionStorage.getItem(STORAGE_TOKEN);
  }

  private loadStoredUser(): AuthUser | null {
    if (typeof sessionStorage === 'undefined') return null;
    const raw = sessionStorage.getItem(STORAGE_USER);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as AuthUser;
    } catch {
      return null;
    }
  }
}
