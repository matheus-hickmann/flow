import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

test.describe('Relatórios', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
    await page.goto('/relatorios');
  });

  test('deve exibir a página de relatórios', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Relatórios' })).toBeVisible();
  });

  test('deve exibir filtros de período e tipo', async ({ page }) => {
    await expect(page.locator('input[type="month"]').first()).toBeVisible();
    await expect(page.locator('select')).toBeVisible();
    await expect(page.getByText('Atualizar')).toBeVisible();
  });

  test('deve carregar dados ao clicar em Atualizar', async ({ page }) => {
    await page.click('button:has-text("Atualizar")');
    // Wait for loading to complete (button text returns to "Atualizar")
    await expect(page.getByText('Atualizar')).toBeVisible({ timeout: 10000 });
  });

  test('deve exibir opções de tipo: Despesas, Receitas, Todos', async ({ page }) => {
    const typeSelect = page.locator('select');
    await expect(typeSelect.locator('option:has-text("Despesas")')).toHaveCount(1);
    await expect(typeSelect.locator('option:has-text("Receitas")')).toHaveCount(1);
    await expect(typeSelect.locator('option:has-text("Todos")')).toHaveCount(1);
  });

  // Business question: Should the chart appear even without data (empty state)?
  // Currently: shows "Nenhum dado para o período selecionado." message.
  // Is this the intended behavior for new users?

  // Business question: Should the report be filtered to show only the user's own
  // custom categories or all categories found in transactions?
  // Currently: shows categories from actual transactions, not the saved category list.
  // These may diverge if user renames categories after recording transactions.
});
