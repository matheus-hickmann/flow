import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

test.describe('Tema escuro', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
  });

  test('deve alternar para tema escuro ao clicar no toggle', async ({ page }) => {
    const toggleBtn = page.locator('button[title="Modo escuro"], button[title="Modo claro"]').first();
    await toggleBtn.click();
    const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
    expect(isDark).toBe(true);
  });

  test('deve persistir preferência de tema após reload', async ({ page }) => {
    const toggleBtn = page.locator('button[title="Modo escuro"], button[title="Modo claro"]').first();
    await toggleBtn.click(); // enable dark
    await page.reload();
    const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
    expect(isDark).toBe(true);
  });

  test('deve voltar ao modo claro ao clicar no toggle novamente', async ({ page }) => {
    const toggleBtn = page.locator('button[title="Modo escuro"], button[title="Modo claro"]').first();
    await toggleBtn.click(); // enable dark
    await toggleBtn.click(); // back to light
    const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
    expect(isDark).toBe(false);
  });
});
