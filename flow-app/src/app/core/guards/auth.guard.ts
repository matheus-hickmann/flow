import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthService } from '../services/auth.service';
import { ENVIRONMENT } from '../config';

/**
 * When apiUrl is set, requires the user to be logged in; otherwise redirects to /login.
 * When apiUrl is not set (mock mode), allows access without login.
 */
export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const env = inject(ENVIRONMENT);
  const router = inject(Router);

  if (!env.authApiUrl && !env.apiUrl) {
    return true;
  }
  if (auth.isLoggedIn()) {
    return true;
  }
  return router.createUrlTree(['/login']);
};
