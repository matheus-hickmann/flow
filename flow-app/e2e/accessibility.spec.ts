import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { signupNew } from './helpers/auth.helper';

// Accessibility tests using axe-core (WCAG 2.1 AA).
// color-contrast is disabled because the Flow design system uses intentional
// pastel tones that do not meet strict WCAG AA contrast ratios by design.

test.describe('Acessibilidade — páginas públicas', () => {
  test('página de login não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/login');
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página de cadastro não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/criar-conta');
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });
});

test.describe('Acessibilidade — páginas autenticadas', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
  });

  test('dashboard não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página de contas não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/contas');
    await expect(page.getByRole('heading', { name: 'Contas' })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página de transações não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/transacoes');
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página de relatórios não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/relatorios');
    await expect(page.getByRole('heading', { name: 'Relatórios' })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página de categorias não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/categorias');
    await expect(page.getByRole('heading', { name: 'Categorias' })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página de dívidas não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/dividas');
    await expect(page.getByRole('heading', { name: /Dívidas/i })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('página família não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/familia');
    await expect(page.getByRole('heading', { name: 'Família' })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('modal de nova dívida não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/dividas');
    await page.getByRole('button', { name: 'Nova dívida' }).click();
    await expect(page.getByText('Registrar')).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test('modal de nova conta não deve ter violações de acessibilidade críticas', async ({ page }) => {
    await page.goto('/contas');
    await page.getByRole('button', { name: 'Nova conta' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    const results = await new AxeBuilder({ page })
      .disableRules(['color-contrast'])
      .analyze();
    expect(results.violations).toEqual([]);
  });
});
