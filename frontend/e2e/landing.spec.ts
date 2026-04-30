import { test, expect } from "@playwright/test";

test("landing page loads with title", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveTitle(/Mini Meeting/);
});

test("landing page has login button", async ({ page }) => {
  await page.goto("/");
  const loginLink = page.getByRole("link", { name: /login/i });
  await expect(loginLink).toBeVisible();
});
