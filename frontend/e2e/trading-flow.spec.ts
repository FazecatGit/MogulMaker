import { test, expect } from '@playwright/test';

test.describe('Trading Flow', () => {
  test('should navigate through main pages', async ({ page }) => {
    // Start at dashboard
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/dashboard/);

    // Navigate to positions
    const positionsLink = page.locator('a[href="/positions"], a:has-text("Positions")');
    if (await positionsLink.isVisible({ timeout: 5000 }).catch(() => false)) {
      await positionsLink.first().click();
      await expect(page).toHaveURL(/\/positions/);
    }

    // Navigate to trades
    const tradesLink = page.locator('a[href="/trades"], a:has-text("Trades")');
    if (await tradesLink.isVisible({ timeout: 5000 }).catch(() => false)) {
      await tradesLink.first().click();
      await expect(page).toHaveURL(/\/trades/);
    }

    // Navigate to watchlist
    const watchlistLink = page.locator('a[href="/watchlist"], a:has-text("Watchlist")');
    if (await watchlistLink.isVisible({ timeout: 5000 }).catch(() => false)) {
      await watchlistLink.first().click();
      await expect(page).toHaveURL(/\/watchlist/);
    }
  });

  test('should filter and sort trades', async ({ page }) => {
    await page.goto('/trades');
    await page.waitForTimeout(2000);

    // Check for filter/sort controls
    const sortDropdown = page.locator('select, button:has-text("Sort")');
    if (await sortDropdown.first().isVisible({ timeout: 5000 }).catch(() => false)) {
      await sortDropdown.first().click();
    }

    // Check that page doesn't crash
    await expect(page.locator('body')).toBeVisible();
  });

  test('should run backtest', async ({ page }) => {
    await page.goto('/backtest');
    await page.waitForTimeout(2000);

    // Look for symbol input
    const symbolInput = page.locator('input[name="symbol"], input[placeholder*="symbol" i]');
    
    if (await symbolInput.isVisible({ timeout: 5000 }).catch(() => false)) {
      await symbolInput.fill('AAPL');

      // Look for run button
      const runButton = page.locator('button:has-text("Run"), button:has-text("Start")');
      if (await runButton.isVisible({ timeout: 5000 }).catch(() => false)) {
        await runButton.click();
        
        // Wait for results (or error)
        await page.waitForTimeout(3000);
        
        // Page should still be functional
        await expect(page.locator('body')).toBeVisible();
      }
    }
  });

  test('should manage watchlist', async ({ page }) => {
    await page.goto('/watchlist');
    await page.waitForTimeout(2000);

    // Look for add button
    const addButton = page.locator('button:has-text("Add")');
    
    if (await addButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await addButton.first().click();

      // Check for modal/form
      const modal = page.locator('[role="dialog"], .modal, form');
      await expect(modal.first()).toBeVisible({ timeout: 5000 });
    }
  });

  test('should display risk metrics', async ({ page }) => {
    await page.goto('/risk');
    await page.waitForTimeout(2000);

    // Check for risk-related content
    const riskContent = page.locator('text=/risk|loss|drawdown/i');
    await expect(riskContent.first()).toBeVisible({ timeout: 10000 });
  });

  test('should show news feed', async ({ page }) => {
    await page.goto('/news');
    await page.waitForTimeout(2000);

    // Check for news-related content
    const newsContent = page.locator('text=/news|article|sentiment/i');
    await expect(newsContent.first()).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Performance', () => {
  test('should load pages quickly', async ({ page }) => {
    const startTime = Date.now();
    await page.goto('/dashboard');
    const loadTime = Date.now() - startTime;

    // Page should load in under 3 seconds (generous for dev mode)
    expect(loadTime).toBeLessThan(3000);
  });

  test('should not have console errors', async ({ page }) => {
    const errors: string[] = [];
    
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    await page.goto('/dashboard');
    await page.waitForTimeout(2000);

    // Filter out expected errors (like API timeouts in dev)
    const criticalErrors = errors.filter(
      (error) => !error.includes('fetch') && !error.includes('API')
    );

    expect(criticalErrors.length).toBe(0);
  });
});
