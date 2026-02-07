import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard');
  });

  test('should load dashboard page', async ({ page }) => {
    await expect(page).toHaveTitle(/Dashboard/);
  });

  test('should display portfolio metrics', async ({ page }) => {
    // Wait for data to load
    await page.waitForTimeout(2000);

    // Check for key dashboard elements
    const portfolioSection = page.locator('text=Portfolio');
    await expect(portfolioSection).toBeVisible({ timeout: 10000 });
  });

  test('should show charts', async ({ page }) => {
    // Wait for charts to render
    await page.waitForTimeout(2000);

    // Recharts creates SVG elements
    const charts = page.locator('svg');
    await expect(charts.first()).toBeVisible({ timeout: 10000 });
  });

  test('should be responsive', async ({ page }) => {
    // Test mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('body')).toBeVisible();

    // Test tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('body')).toBeVisible();

    // Test desktop viewport
    await page.setViewportSize({ width: 1920, height: 1080 });
    await expect(page.locator('body')).toBeVisible();
  });
});
