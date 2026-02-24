import { InjectionToken } from '@angular/core';

import type { Environment } from './environment.model';

export const ENVIRONMENT = new InjectionToken<Environment>('ENVIRONMENT');
