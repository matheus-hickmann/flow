import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

// NOTE: tests that verify a debt appears in the list after creation require the
// backend running at http://localhost:8080. Tests that only exercise UI state
// (modals, form validation) work without a backend.

test.describe('Dívidas', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
    await page.goto('/dividas');
  });

  test('deve exibir a página de dívidas', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Dívidas/i })).toBeVisible();
  });

  test('deve exibir estado vazio quando não há dívidas', async ({ page }) => {
    // A fresh account has no debts — empty state message should be visible.
    await expect(page.getByText('Nenhuma dívida registrada.')).toBeVisible();
  });

  test('deve abrir modal de nova dívida ao clicar em Nova dívida', async ({ page }) => {
    await page.getByRole('button', { name: 'Nova dívida' }).click();
    await expect(page.getByText('NOVA DÍVIDA')).toBeVisible();
    await expect(page.getByText('Registrar')).toBeVisible();
  });

  test('deve exibir toggle de tipo TO_PAY e TO_RECEIVE', async ({ page }) => {
    await page.getByRole('button', { name: 'Nova dívida' }).click();
    await expect(page.getByRole('button', { name: 'Devo a alguém' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Me devem' })).toBeVisible();
  });

  test('deve bloquear salvar dívida sem campos obrigatórios', async ({ page }) => {
    await page.getByRole('button', { name: 'Nova dívida' }).click();
    // Click Salvar without filling required fields
    await page.getByRole('button', { name: 'Salvar' }).click();
    // Form error should appear and modal should remain open
    await expect(page.getByText('Nome e valor são obrigatórios.')).toBeVisible();
    await expect(page.getByText('NOVA DÍVIDA')).toBeVisible();
  });

  // Requires backend. Skipped when running without the Go service.
  test('deve criar dívida do tipo "a pagar" e exibi-la na lista', async ({ page }) => {
    test.fixme(true, 'Requires Go backend running at http://localhost:8080');
    const name = `Empréstimo-${Date.now()}`;
    await page.getByRole('button', { name: 'Nova dívida' }).click();
    await page.getByRole('button', { name: 'Devo a alguém' }).click();
    await page.getByPlaceholder('ex: Empréstimo pessoal').fill(name);
    await page.getByPlaceholder('0,00').fill('500');
    await page.getByPlaceholder('Nome da pessoa ou empresa').fill('Banco XYZ');
    await page.getByRole('button', { name: 'Salvar' }).click();
    // After saving, modal closes and debt appears in the active list
    await expect(page.getByText('NOVA DÍVIDA')).not.toBeVisible({ timeout: 8000 });
    await expect(page.getByText(name)).toBeVisible();
    await expect(page.getByText('DÍVIDAS ATIVAS')).toBeVisible();
  });

  test('deve abrir modal de pagamento ao clicar em Registrar pagamento', async ({ page }) => {
    test.fixme(true, 'Requires a pre-existing debt — needs Go backend at http://localhost:8080');
    // This flow only works when there is at least one active debt in the list.
    await page.getByRole('button', { name: 'Registrar pagamento' }).first().click();
    await expect(page.getByText('PAGAMENTO')).toBeVisible();
  });

  test('deve fechar modal ao clicar em Cancelar', async ({ page }) => {
    await page.getByRole('button', { name: 'Nova dívida' }).click();
    await expect(page.getByText('NOVA DÍVIDA')).toBeVisible();
    await page.getByRole('button', { name: 'Cancelar' }).click();
    await expect(page.getByText('NOVA DÍVIDA')).not.toBeVisible();
  });
});
