import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { ROUTES } from '../../core/constants/app.constants';
import { ENVIRONMENT } from '../../core/config';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './header.component.html',
})
export class HeaderComponent {
  protected readonly routes = ROUTES;
  protected readonly appName = inject(ENVIRONMENT).appName;
  protected readonly userName = inject(ENVIRONMENT).defaultUserName;
}
