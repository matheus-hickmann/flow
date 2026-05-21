import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

test.describe('Lançamentos', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
  });

  test('deve exibir o FAB menu ao clicar no botão +', async ({ page }) => {
    const fabBtn = page.locator('button[aria-label*="Abrir ou fechar"]');
    await fabBtn.click();
    await expect(page.getByText('Despesa')).toBeVisible();
    await expect(page.getByText('Receita')).toBeVisible();
    await expect(page.getByText('Transferência')).toBeVisible();
    await expect(page.getByText('Planejamento')).toBeVisible();
  });

  test('deve abrir modal de despesa ao clicar em Despesa no FAB', async ({ page }) => {
    await page.locator('button[aria-label*="Abrir ou fechar"]').click();
    await page.getByText('Despesa').click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Nova Despesa')).toBeVisible();
  });

  test('deve abrir modal de receita ao clicar em Receita no FAB', async ({ page }) => {
    await page.locator('button[aria-label*="Abrir ou fechar"]').click();
    await page.getByText('Receita').click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Nova Receita')).toBeVisible();
  });

  test('deve bloquear submit de despesa sem preencher campos obrigatórios', async ({ page }) => {
    await page.locator('button[aria-label*="Abrir ou fechar"]').click();
    await page.getByText('Despesa').click();
    await page.click('button[type="submit"]');
    // form should still be visible (not submitted)
    await expect(page.getByRole('dialog')).toBeVisible();
  });

  test('deve fechar modal de despesa ao clicar em Cancelar', async ({ page }) => {
    await page.locator('button[aria-label*="Abrir ou fechar"]').click();
    await page.getByText('Despesa').click();
    await page.click('button:has-text("Cancelar")');
    await expect(page.getByRole('dialog')).not.toBeVisible();
  });

  // Business question: After posting a transaction, should it appear immediately in
  // the transaction list on the same page without a full reload?
  // Assumption: yes — transactionService.refresh() triggers this.

  // Business question: Is the "Planejamento" field in expense/income modal only shown
  // when there are existing budget limits for the user? Currently: yes (conditional render).
  // This is correct behavior but needs confirmation.
});
