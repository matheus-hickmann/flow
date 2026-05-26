import { Component, input, output } from '@angular/core';

@Component({
  selector: 'app-modal',
  standalone: true,
  templateUrl: './modal.component.html',
})
export class ModalComponent {
  title = input.required<string>();
  eyebrow = input<string>('');
  closeOnOverlay = input<boolean>(true);

  readonly closed = output<void>();

  close(): void {
    this.closed.emit();
  }

  onOverlayClick(event: Event): void {
    if (this.closeOnOverlay() && (event.target as HTMLElement).classList.contains('modal-overlay')) {
      this.close();
    }
  }
}
