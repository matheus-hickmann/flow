import { ChangeDetectionStrategy, Component, Input } from '@angular/core';

@Component({
  selector: 'flow-mark',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    @if (variant === 'wordmark') {
      <svg
        [attr.height]="height"
        [attr.width]="height * (168 / 56)"
        viewBox="0 0 168 56"
        role="img"
        aria-label="Flow"
        style="display:block;overflow:visible">
        <text
          x="0" y="40"
          font-family="'Instrument Serif','Times New Roman',serif"
          font-style="italic" font-weight="400"
          font-size="52" letter-spacing="-1.2"
          fill="currentColor">flow</text>
        <g class="flow-mark__accent">
          <path d="M 4 52 C 28 55, 62 55, 92 52 C 122 49, 142 38, 158 26"
                fill="none" stroke="currentColor" stroke-width="1.5"
                stroke-linecap="round" />
          <circle cx="4"   cy="52" r="1.8" fill="currentColor" />
          <circle cx="92"  cy="52" r="1.8" fill="currentColor" />
          <circle cx="158" cy="26" r="1.8" fill="currentColor" />
        </g>
      </svg>
    }
    @if (variant === 'icon') {
      <svg
        [attr.height]="height"
        [attr.width]="height"
        viewBox="0 0 56 56"
        role="img"
        aria-label="Flow"
        style="display:block">
        <g class="flow-mark__accent">
          <path d="M 6 44 C 18 47, 28 46, 36 40 C 44 34, 48 22, 50 12"
                fill="none" stroke="currentColor" stroke-width="3"
                stroke-linecap="round" />
          <circle cx="6"  cy="44" r="3"   fill="currentColor" />
          <circle cx="36" cy="40" r="3"   fill="currentColor" />
          <circle cx="50" cy="12" r="3.5" fill="currentColor" />
        </g>
      </svg>
    }
  `,
  styles: [`
    :host { display: inline-block; line-height: 0; }
  `],
})
export class FlowMarkComponent {
  @Input() height = 32;
  @Input() variant: 'wordmark' | 'icon' = 'wordmark';
}
