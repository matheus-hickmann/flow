import { Page } from '@playwright/test';

export async function loginAs(page: Page, userId: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.fill('#userId', userId);
  await page.fill('#password', password);
  await page.click('button[type="submit"]');
  await page.waitForURL('/');
}

export async function signupNew(page: Page): Promise<{ userId: string; password: string }> {
  await page.goto('/criar-conta');
  // read auto-generated userId from read-only field
  const userId = await page.inputValue('#userId');
  const password = `Senha@${Date.now()}`;
  await page.fill('#displayName', 'Teste E2E');
  await page.fill('#password', password);

  // fill recovery questions using index-based selection (index 0 is the disabled placeholder)
  const selects = page.locator('select[name^="question"]');
  const inputs = page.locator('input[name^="answer"]');
  for (let i = 0; i < 3; i++) {
    await selects.nth(i).selectOption({ index: i + 1 });
    await inputs.nth(i).fill(`resposta${i + 1}`);
  }

  await page.click('button[type="submit"]');
  await page.waitForURL('/');
  return { userId, password };
}
