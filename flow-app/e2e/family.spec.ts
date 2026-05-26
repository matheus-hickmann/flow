import { test, expect } from '@playwright/test';
import { signupNew } from './helpers/auth.helper';

// NOTE: test that verifies a group appears in the list after creation requires
// the Go backend running at http://localhost:8080.

test.describe('Família', () => {
  test.beforeEach(async ({ page }) => {
    await signupNew(page);
    await page.goto('/familia');
  });

  test('deve exibir a página família', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Família' })).toBeVisible();
  });

  test('deve exibir botão Novo grupo', async ({ page }) => {
    await expect(page.getByRole('button', { name: '+ Novo grupo' })).toBeVisible();
  });

  test('deve exibir estado vazio para novo usuário', async ({ page }) => {
    // A freshly created account has no groups, so the empty state should render.
    // The template shows this message when groups().length === 0 (after loading).
    await expect(
      page.getByText('Você ainda não faz parte de nenhum grupo familiar.'),
    ).toBeVisible({ timeout: 8000 });
  });

  test('deve abrir formulário de criação de grupo ao clicar em Novo grupo', async ({ page }) => {
    await page.getByRole('button', { name: '+ Novo grupo' }).click();
    await expect(page.getByRole('heading', { name: 'Novo grupo familiar' })).toBeVisible();
    await expect(page.getByPlaceholder('Ex: Família Silva')).toBeVisible();
  });

  test('deve fechar modal de criação de grupo ao clicar em Cancelar', async ({ page }) => {
    await page.getByRole('button', { name: '+ Novo grupo' }).click();
    await expect(page.getByRole('heading', { name: 'Novo grupo familiar' })).toBeVisible();
    await page.getByRole('button', { name: 'Cancelar' }).click();
    await expect(page.getByRole('heading', { name: 'Novo grupo familiar' })).not.toBeVisible();
  });

  test('deve exibir erro ao tentar criar grupo com nome vazio', async ({ page }) => {
    await page.getByRole('button', { name: '+ Novo grupo' }).click();
    // Submit without filling the name
    await page.getByRole('button', { name: 'Criar' }).click();
    await expect(page.getByText('Informe um nome para o grupo.')).toBeVisible();
  });

  // Requires backend. Skipped when running without the Go service.
  test('deve criar um grupo e exibi-lo na lista', async ({ page }) => {
    test.fixme(true, 'Requires Go backend running at http://localhost:8080');
    const groupName = `Família-${Date.now()}`;
    await page.getByRole('button', { name: '+ Novo grupo' }).click();
    await page.getByPlaceholder('Ex: Família Silva').fill(groupName);
    await page.getByRole('button', { name: 'Criar' }).click();
    // Modal should close and group should appear in the sidebar list
    await expect(page.getByRole('heading', { name: 'Novo grupo familiar' })).not.toBeVisible({
      timeout: 8000,
    });
    await expect(page.getByText(groupName)).toBeVisible();
    // Group detail panel should also show the name
    await expect(page.getByRole('heading', { name: groupName })).toBeVisible();
  });
});
