import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
    // signupNew already lands on '/', but navigate explicitly for clarity
    await page.goto('/');
  });

  test('deve exibir a página do dashboard', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });

  test('deve exibir navegação de mês', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Mês anterior' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Próximo mês' })).toBeVisible();
  });

  test('deve navegar para o mês anterior', async ({ page }) => {
    // Capture current month label, navigate back and verify label changes
    const heading = page.getByRole('heading', { name: 'Dashboard' });
    await expect(heading).toBeVisible();
    const before = await page.locator('h1').textContent();
    await page.getByRole('button', { name: 'Mês anterior' }).click();
    // Month label embedded in <h1> should update after navigation
    await expect(page.locator('h1')).not.toHaveText(before ?? '');
  });

  test('deve navegar para o próximo mês', async ({ page }) => {
    const heading = page.getByRole('heading', { name: 'Dashboard' });
    await expect(heading).toBeVisible();
    const before = await page.locator('h1').textContent();
    await page.getByRole('button', { name: 'Próximo mês' }).click();
    await expect(page.locator('h1')).not.toHaveText(before ?? '');
  });

  test('deve exibir seção de últimos lançamentos', async ({ page }) => {
    await expect(page.getByText('Últimos lançamentos')).toBeVisible();
  });

  test('deve exibir seção de saldo em contas', async ({ page }) => {
    await expect(page.getByText('Saldo em contas')).toBeVisible();
  });

  test('deve exibir botão Exportar', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Exportar' })).toBeVisible();
  });
});
