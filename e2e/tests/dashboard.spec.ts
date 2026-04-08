import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('should display login page', async ({ page }) => {
    await page.goto('/login');
    await expect(page).toHaveTitle(/AnubisWatch|Login/);
    await expect(page.getByPlaceholder(/username|email/i)).toBeVisible();
    await expect(page.getByPlaceholder(/password/i)).toBeVisible();
  });
});

test.describe('Dashboard', () => {
  test('should display navigation', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Anubis')).toBeVisible();
    await expect(page.getByText('Dashboard')).toBeVisible();
    await expect(page.getByText('Souls')).toBeVisible();
  });
});

test.describe('API', () => {
  test('health endpoint should return ok', async ({ request }) => {
    const response = await request.get('/api/health');
    expect(response.ok()).toBeTruthy();
  });
});
