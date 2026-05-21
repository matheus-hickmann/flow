import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

test.describe('Categorias', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
    await page.goto('/categorias');
  });

  test('deve exibir a página de categorias', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Categorias' })).toBeVisible();
  });

  test('deve exibir as abas Despesas e Receitas', async ({ page }) => {
    await expect(page.getByText('Despesas')).toBeVisible();
    await expect(page.getByText('Receitas')).toBeVisible();
  });

  test('deve exibir categorias padrão de despesas', async ({ page }) => {
    await expect(page.getByText('Alimentação')).toBeVisible();
    await expect(page.getByText('Moradia')).toBeVisible();
  });

  test('deve adicionar uma nova categoria de despesa', async ({ page }) => {
    await page.fill('input[placeholder="Ex.: Streaming"]', 'Streaming');
    await page.click('button:has-text("Adicionar")');
    await expect(page.getByText('Streaming')).toBeVisible();
  });

  test('deve remover uma categoria', async ({ page }) => {
    const count = await page.locator('button[title="Remover"]').count();
    await page.locator('button[title="Remover"]').first().click();
    const newCount = await page.locator('button[title="Remover"]').count();
    expect(newCount).toBe(count - 1);
  });

  test('deve trocar para aba de Receitas', async ({ page }) => {
    await page.click('button:has-text("Receitas")');
    await expect(page.getByText('Salário')).toBeVisible();
  });

  // Business question: Should removing all categories be allowed, or should there be
  // a minimum of 1 category? Currently no minimum enforced — needs decision.
});
