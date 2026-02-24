import { Component, input, output } from '@angular/core';
import { animate, style, transition, trigger } from '@angular/animations';

export type FabOptionId = 'expense' | 'income' | 'planning';

export interface FabOption {
  readonly id: FabOptionId;
  readonly label: string;
}

@Component({
  selector: 'app-fab-menu',
  standalone: true,
  animations: [
    trigger('fabExpand', [
      transition(':enter', [
        style({ opacity: 0, transform: 'scale(0.85) translateY(6px)' }),
        animate('220ms ease-out', style({ opacity: 1, transform: 'scale(1) translateY(0)' })),
      ]),
      transition(':leave', [
        animate('160ms ease-in', style({ opacity: 0, transform: 'scale(0.92) translateY(4px)' })),
      ]),
    ]),
  ],
  templateUrl: './fab-menu.component.html',
})
export class FabMenuComponent {
  isOpen = input.required<boolean>();
  options = input.required<readonly FabOption[]>();
  ariaLabelToggle = input.required<string>();

  readonly openChange = output<void>();
  readonly optionSelect = output<FabOptionId>();

  onOptionClick(optionId: FabOptionId): void {
    this.optionSelect.emit(optionId);
  }
}
