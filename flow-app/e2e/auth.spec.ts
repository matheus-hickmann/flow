import { test, expect } from '@playwright/test';
import { signupNew, loginAs } from './helpers/auth.helper';

test.describe('Autenticação', () => {
  test('deve exibir a tela de login ao acessar rota protegida sem autenticação', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL('/login');
  });

  test('deve exibir erro ao tentar logar com credenciais inválidas', async ({ page }) => {
    await page.goto('/login');
    await page.fill('#userId', 'INVALIDO123456');
    await page.fill('#password', 'senha_errada');
    await page.click('button[type="submit"]');
    await expect(page.locator('[class*="red"]').first()).toBeVisible();
  });

  test('deve gerar um userId de 12 caracteres na tela de cadastro', async ({ page }) => {
    await page.goto('/criar-conta');
    const userId = await page.inputValue('#userId');
    expect(userId).toHaveLength(12);
    expect(userId).toMatch(/^[A-Z0-9]+$/);
  });

  test('deve permitir regerar o userId na tela de cadastro', async ({ page }) => {
    await page.goto('/criar-conta');
    const first = await page.inputValue('#userId');
    await page.click('button[title="Gerar novo ID"]');
    const second = await page.inputValue('#userId');
    expect(first).not.toBe(second);
  });

  test('deve exibir aviso para salvar o userId', async ({ page }) => {
    await page.goto('/criar-conta');
    await expect(page.getByText('Guarde este ID em um local seguro')).toBeVisible();
  });

  // Business question: Does signup require all 3 recovery questions?
  // Assumption: yes — form validates presence of all 3.
  test('deve bloquear cadastro sem preencher todas as perguntas de recuperação', async ({ page }) => {
    await page.goto('/criar-conta');
    await page.fill('#password', 'Senha@123');
    await page.click('button[type="submit"]');
    await expect(page.locator('[class*="red"]').first()).toBeVisible();
  });
});
