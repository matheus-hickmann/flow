import { Pipe, PipeTransform } from '@angular/core';

/**
 * Formata valor numérico como moeda BRL (R$). Use em templates para consistência.
 */
@Pipe({ name: 'currencyBrl', standalone: true })
export class CurrencyBrlPipe implements PipeTransform {
  transform(value: number | null | undefined): string {
    if (value == null || Number.isNaN(value)) {
      return 'R$ 0,00';
    }
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
  }
}
