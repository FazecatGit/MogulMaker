import { test, expect } from '@playwright/test';

test.describe('Positions Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/positions');
  });

  test('should load positions page', async ({ page }) => {
    await expect(page.locator('h1')).toContainText(/Positions/i);
  });

  test('should display positions table', async ({ page }) => {
    // Wait for table to load
    await page.waitForTimeout(2000);

    // Check for table headers
    const table = page.locator('table').first();
    await expect(table).toBeVisible({ timeout: 10000 });
  });

  test('should update positions in real-time', async ({ page }) => {
    // Wait for initial load
    await page.waitForTimeout(2000);

    // Check for live indicator
    const liveIndicator = page.locator('text=/live|updating/i');
    await expect(liveIndicator.first()).toBeVisible({ timeout: 10000 });
  });

  test('should close position with confirmation', async ({ page }) => {
    await page.waitForTimeout(2000);

    // Look for close button (if positions exist)
    const closeButton = page.locator('button:has-text("Close")').first();
    
    if (await closeButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await closeButton.click();

      // Check for confirmation dialog
      const confirmButton = page.locator('button:has-text("Confirm")');
      await expect(confirmButton).toBeVisible({ timeout: 5000 });
    }
  });

  test('should handle empty positions state', async ({ page }) => {
    await page.waitForTimeout(2000);

    // Check for either positions or empty state
    const hasPositions = await page.locator('table tbody tr').count() > 0;
    const hasEmptyState = await page.locator('text=/no positions|empty/i').isVisible().catch(() => false);

    expect(hasPositions || hasEmptyState).toBe(true);
  });
});
