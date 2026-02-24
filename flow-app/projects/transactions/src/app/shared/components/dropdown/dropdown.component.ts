import { Component, input, output } from '@angular/core';

@Component({
  selector: 'app-dropdown',
  standalone: true,
  templateUrl: './dropdown.component.html',
  styles: [`
    .select-dropdown-icon {
      background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2364748b'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E");
      background-size: 1rem 1rem;
      background-position: right 0.875rem center;
      padding-right: 1.875rem;
    }
  `],
})
export class DropdownComponent {
  id = input<string>('');
  value = input<string>('');
  ariaLabel = input<string | undefined>(undefined);

  readonly valueChange = output<string>();

  onChange(event: Event): void {
    const value = (event.target as HTMLSelectElement).value;
    this.valueChange.emit(value);
  }
}
