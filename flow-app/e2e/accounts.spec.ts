import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

// NOTE: these tests require the backend running at http://localhost:8080
// and flow-app at http://localhost:4200

test.describe('Contas', () => {
  test.beforeEach(async ({ page }) => {
    // Use a fresh signup for isolation
    // In CI, consider using a fixed test user seeded in DynamoDB Local
    await signupNew(page);
    await page.goto('/contas');
  });

  test('deve exibir a página de contas sem erro', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Contas' })).toBeVisible();
  });

  test('deve abrir modal de nova conta', async ({ page }) => {
    await page.click('button:has-text("Nova conta")');
    await expect(page.getByRole('dialog')).toBeVisible();
  });

  test('deve mostrar os 3 tipos de conta no modal', async ({ page }) => {
    await page.click('button:has-text("Nova conta")');
    await expect(page.getByText('Corrente / Poupança')).toBeVisible();
    await expect(page.getByText('Investimento')).toBeVisible();
    await expect(page.getByText('Cartão de crédito')).toBeVisible();
  });

  test('deve exibir campos de cartão ao selecionar tipo Cartão de crédito', async ({ page }) => {
    await page.click('button:has-text("Nova conta")');
    await page.click('label:has-text("Cartão de crédito") input');
    await expect(page.locator('#acc-brand')).toBeVisible();
    await expect(page.locator('#acc-limit')).toBeVisible();
    await expect(page.locator('#acc-closing')).toBeVisible();
    await expect(page.locator('#acc-due')).toBeVisible();
  });

  test('deve ocultar campo de saldo inicial para cartão de crédito', async ({ page }) => {
    await page.click('button:has-text("Nova conta")');
    await page.click('label:has-text("Cartão de crédito") input');
    await expect(page.locator('#acc-initial')).not.toBeVisible();
  });

  // Business question: Should account name be unique per user?
  // Skipping this test until confirmed.
});
